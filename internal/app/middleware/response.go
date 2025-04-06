package middleware

import (
	"context"
	"log"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// WithResponse creates middleware that handles writing the response
func WithResponse[T any](handler func(ctx context.Context, req T) (interface{}, error)) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// Store request context for handlers that need it
		ctx = context.WithValue(ctx, "requestContext", c)

		var req T
		if v, exists := c.Get("parsedRequest"); exists {
			req = v.(T)
		}

		resp, err := handler(ctx, req)
		if err != nil {
			// Try to convert to APIError
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
