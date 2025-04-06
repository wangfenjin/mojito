package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// WithRequest creates middleware that parses the request
func WithRequest[T any](requestType T) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		var req T
		if err := c.BindAndValidate(&req); err != nil {
			c.JSON(consts.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
			c.Abort()
			return
		}
		
		// Store the parsed request in the context
		c.Set("parsedRequest", req)
		c.Next(ctx)
	}
}
