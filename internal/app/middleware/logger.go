package middleware

import (
	"github.com/gin-gonic/gin"
)

// LoggerMiddleware logs HTTP request/response details
func LoggerMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO
		ctx.Next()
	}
}
