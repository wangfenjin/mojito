package middleware

import (
	"context"
	"log"
	"reflect"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// WithHandler creates middleware that handles both request parsing and response writing
func WithHandler[Req any, Resp any](handler func(ctx context.Context, req Req) (Resp, error)) app.HandlerFunc {
	// Register route information at middleware creation time, not request time
	// Get type information using reflection
	reqType := reflect.TypeOf(*new(Req))
	respType := reflect.TypeOf(*new(Resp))

	// Store handler information for documentation
	handlerInfo := HandlerInfo{
		Handler:      handler,
		RequestType:  reqType,
		ResponseType: respType,
	}

	// Add to pending handlers that will be registered when routes are set up
	AddPendingHandler(handlerInfo)

	return func(ctx context.Context, c *app.RequestContext) {
		// Store request context for handlers that need it
		ctx = context.WithValue(ctx, "requestContext", c)
		var req Req
		if err := c.BindAndValidate(&req); err != nil {
			AbortWithError(c, NewBadRequestError(err.Error()))
			return
		}

		resp, err := handler(ctx, req)
		if err != nil {
			if apiErr, ok := err.(*APIError); ok {
				log.Printf("API Error: %v, Code: %d, Path: %s, Method: %s",
					apiErr.Message, apiErr.Code, c.Path(), c.Method())
				c.JSON(apiErr.Code, map[string]interface{}{
					"error": apiErr.Message,
				})
			} else {
				log.Printf("Internal Server Error: %v, Path: %s, Method: %s",
					err, c.Path(), c.Method())
				c.JSON(consts.StatusInternalServerError, map[string]interface{}{
					"error": err.Error(),
				})
			}
			return
		}

		c.JSON(consts.StatusOK, resp)
	}
}

// WithHandlerEmpty is a convenience function for handlers without request body
func WithHandlerEmpty[Resp any](handler func(ctx context.Context) (Resp, error)) app.HandlerFunc {
	// Register empty request handler
	respType := reflect.TypeOf(*new(Resp))

	// Store handler information for documentation
	handlerInfo := HandlerInfo{
		Handler:      handler,
		RequestType:  nil, // No request type
		ResponseType: respType,
	}

	// Add to pending handlers
	AddPendingHandler(handlerInfo)

	return WithHandler(func(ctx context.Context, _ struct{}) (Resp, error) {
		return handler(ctx)
	})
}
