package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// WithResponse creates middleware that handles writing the response
func WithResponse[T any](handler func(ctx context.Context, req T) (interface{}, error)) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		var req T
		if v, exists := c.Get("parsedRequest"); exists {
			req = v.(T)
		}

		// Call the handler with the parsed request
		resp, err := handler(ctx, req)
		if err != nil {
			if badReqErr, ok := err.(*BadRequestError); ok {
				c.JSON(badReqErr.GetStatusCode(), map[string]interface{}{
					"error": badReqErr.Error(),
				})
				return
			}

			c.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		c.JSON(consts.StatusOK, resp)
	}
}
