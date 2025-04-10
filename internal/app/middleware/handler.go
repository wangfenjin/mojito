package middleware

import (
	"context"
	"net/http"
	"reflect"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/wangfenjin/mojito/internal/app/utils"
	"github.com/wangfenjin/mojito/internal/pkg/logger"
	"github.com/wangfenjin/mojito/pkg/openapi"
)

// WithHandler creates middleware that handles both request parsing and response writing
func WithHandler[Req any, Resp any](handler func(ctx context.Context, req Req) (Resp, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !openapi.Registered(c.Request.Method, c.FullPath()) {
			ms := c.HandlerNames()
			handlerName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
			openapi.RegisterHandler(c.Request.Method, c.FullPath(), handlerName, nil, reflect.TypeOf((*Resp)(nil)).Elem(), ms...)
			logger.GetLogger().Info("Registering handler", "name", handlerName, "path", string(c.FullPath()), "method", string(c.Request.Method), "middleware", ms)
		}

		var req Req
		c.ShouldBind(&req)
		c.ShouldBindUri(&req)
		c.ShouldBindHeader(&req)
		c.ShouldBindQuery(&req)
		if err := binding.Validator.ValidateStruct(req); err != nil {
			logger.GetLogger().Error("Bind error", "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, NewBadRequestError(err.Error()))
			return
		}

		resp, err := handler(c, req)
		if err != nil {
			logger.GetLogger().Error("Handler error", "error", err)
			if apiErr, ok := err.(*APIError); ok {
				c.AbortWithStatusJSON(http.StatusBadRequest, apiErr)
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, NewBadRequestError(err.Error()))
			}
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, NewUnauthorizedError("Authorization header is required"))
			return
		}
		// Extract token from "Bearer <token>"
		if len(token) < 7 || token[:7] != "Bearer " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, NewUnauthorizedError("Invalid Authorization header"))
			return
		}
		token = token[7:]

		claims, err := utils.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, NewUnauthorizedError(err.Error()))
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}
