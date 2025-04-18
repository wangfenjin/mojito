// Package middleware provides HTTP middleware functions for the application
package middleware

import (
	"fmt"
	"net/http"
)

// APIError represents an error that can be returned by the API
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error implement error interface for APIError
func (e *APIError) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string) *APIError {
	return &APIError{
		Code:    http.StatusUnauthorized,
		Message: message,
	}
}

// NewBadRequestError creates a new bad request error
func NewBadRequestError(message string) *APIError {
	return &APIError{
		Code:    http.StatusBadRequest,
		Message: message,
	}
}

// NewForbiddenError creates a new forbidden error
func NewForbiddenError(message string) *APIError {
	return &APIError{
		Code:    http.StatusForbidden,
		Message: message,
	}
}
