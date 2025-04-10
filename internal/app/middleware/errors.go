package middleware

import (
	"fmt"
	"net/http"
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
		Code:    http.StatusUnauthorized,
		Message: message,
	}
}

func NewBadRequestError(message string) *APIError {
	return &APIError{
		Code:    http.StatusBadRequest,
		Message: message,
	}
}

func NewForbiddenError(message string) *APIError {
	return &APIError{
		Code:    http.StatusForbidden,
		Message: message,
	}
}
