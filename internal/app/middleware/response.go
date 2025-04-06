package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
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
			switch e := err.(type) {
			case *BadRequestError:
				c.JSON(e.GetStatusCode(), map[string]interface{}{
					"error": e.Error(),
				})
			case *UnauthorizedError:
				c.JSON(e.GetStatusCode(), map[string]interface{}{
					"error": e.Error(),
				})
			default:
				c.JSON(500, map[string]interface{}{
					"error": err.Error(),
				})
			}
			return
		}

		c.JSON(200, resp)
	}
}
