package middleware

import (
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Implement error interface for APIError
func (e *APIError) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

func NewUnauthorizedError(message string) *APIError {
	return &APIError{
		Code:    consts.StatusUnauthorized,
		Message: message,
	}
}

func NewBadRequestError(message string) *APIError {
	return &APIError{
		Code:    consts.StatusBadRequest,
		Message: message,
	}
}

func NewForbiddenError(message string) *APIError {
	return &APIError{
		Code:    consts.StatusForbidden,
		Message: message,
	}
}

func AbortWithError(c *app.RequestContext, err *APIError) {
	c.JSON(err.Code, map[string]interface{}{
		"error": err.Message,
	})
	c.Abort()
}
