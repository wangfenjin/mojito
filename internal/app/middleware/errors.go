package middleware

import "github.com/cloudwego/hertz/pkg/protocol/consts"

// BadRequestError represents a 400 error
type BadRequestError struct {
	Message string
}

func (e *BadRequestError) Error() string {
	return e.Message
}

// NewBadRequestError creates a new BadRequestError
func NewBadRequestError(message string) *BadRequestError {
	return &BadRequestError{Message: message}
}

// GetStatusCode returns the HTTP status code
func (e *BadRequestError) GetStatusCode() int {
	return consts.StatusBadRequest
}
