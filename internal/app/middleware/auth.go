package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/wangfenjin/mojito/internal/app/utils"
)

func RequireAuth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// Get token from Authorization header
		authHeader := string(c.GetHeader("Authorization"))
		if len(authHeader) == 0 {
			AbortWithError(c, NewUnauthorizedError("missing authorization header"))
			return
		}

		// Extract token from "Bearer <token>"
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			AbortWithError(c, NewUnauthorizedError("invalid authorization format"))
			return
		}
		token := authHeader[7:]

		// Validate token
		claims, err := utils.ValidateToken(token)
		if err != nil {
			AbortWithError(c, NewUnauthorizedError("invalid token"))
			return
		}

		// Store user info in the context that will be passed to handlers
		newCtx := context.WithValue(ctx, "user_id", claims.UserID)
		newCtx = context.WithValue(newCtx, "user_email", claims.Email)

		c.Next(newCtx)
	}
}
