package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

// WithRequest creates middleware that parses the request
func WithRequest[T any](requestType T) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		var req T
		if err := c.BindAndValidate(&req); err != nil {
			AbortWithError(c, NewBadRequestError(err.Error()))
			return
		}

		// Store the parsed request in the context
		c.Set("parsedRequest", req)
		c.Next(ctx)
	}
}
