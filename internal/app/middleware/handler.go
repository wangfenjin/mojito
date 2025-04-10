package middleware

import (
	"context"
	"net/http"
	"reflect"
	"runtime"

	"github.com/gin-gonic/gin"
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
		if err := c.BindUri(&req); err != nil {
			logger.GetLogger().Error("BindUri error", "error", err.Error())
			return
		}
		if err := c.BindHeader(&req); err != nil {
			logger.GetLogger().Error("BindHeader error", "error", err.Error())
			return
		}
		if err := c.BindQuery(&req); err != nil {
			logger.GetLogger().Error("BindQuery error", "error", err.Error())
			return
		}
		if err := c.Bind(&req); err != nil {
			logger.GetLogger().Error("Bind error", "error", err.Error())
			return
		}

		resp, err := handler(c, req)
		if err != nil {
			if apiErr, ok := err.(*APIError); ok {
				c.AbortWithError(apiErr.Code, apiErr).SetType(gin.ErrorTypePublic)
			} else {
				c.AbortWithError(http.StatusInternalServerError, apiErr).SetType(gin.ErrorTypePrivate)
			}
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

// WithHandlerEmpty is a convenience function for handlers without request body
// func WithHandlerEmpty[Resp any](handler func(ctx context.Context) (Resp, error)) app.HandlerFunc {
// 	handlerName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
// 	return WithHandler(func(ctx context.Context, _ any) (Resp, error) {
// 		ctx = context.WithValue(ctx, "handler_name", handlerName)
// 		ctx = context.WithValue(ctx, "request_type", nil)
// 		ctx = context.WithValue(ctx, "response_type", reflect.TypeOf((*Resp)(nil)).Elem())
// 		return handler(ctx)
// 	})
// }

// return func(ctx context.Context, c *app.RequestContext) {
// 	if !openapi.Registered(string(c.Method()), string(c.FullPath())) {
// 		ms := make([]string, 0)
// 		for _, h := range c.Handlers() {
// 			ms = append(ms, runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name())
// 		}
// 		handlerName, ok := ctx.Value("handler_name").(string)
// 		if !ok {
// 			handlerName = runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
// 		}
// 		openapi.RegisterHandler(string(c.Method()), string(c.FullPath()), handlerName, nil, reflect.TypeOf((*Resp)(nil)).Elem(), ms...)
// 		logger.GetLogger().Info("Registering handler", "name", handlerName, "path", string(c.FullPath()), "method", string(c.Method()))
// 	}

// 	// Store request context for handlers that need it
// 	ctx = context.WithValue(ctx, "requestContext", c)
// 	// var req Req
// 	// if err := c.BindAndValidate(&req); err != nil {
// 	// 	AbortWithError(c, NewBadRequestError(err.Error()))
// 	// 	return
// 	// }

// 	resp, err := handler(ctx)
// 	if err != nil {
// 		if apiErr, ok := err.(*APIError); ok {
// 			logger.GetLogger().Error("API Error: ", "message", apiErr.Message, "code", apiErr.Code, "path", c.Path(), "method", c.Method())
// 			c.JSON(apiErr.Code, map[string]interface{}{
// 				"error": apiErr.Message,
// 			})
// 		} else {
// 			logger.GetLogger().Error("Internal Server Error", "path", c.Path(), "method", c.Method())
// 			c.JSON(consts.StatusInternalServerError, map[string]interface{}{
// 				"error": err.Error(),
// 			})
// 		}
// 		return
// 	}

// 	c.JSON(consts.StatusOK, resp)
// }
