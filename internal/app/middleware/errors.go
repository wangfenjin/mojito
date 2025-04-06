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


// UnauthorizedError represents a 401 error
type UnauthorizedError struct {
	Message string
}

func (e *UnauthorizedError) Error() string {
	return e.Message
}

// NewUnauthorizedError creates a new UnauthorizedError
func NewUnauthorizedError(message string) *UnauthorizedError {
	return &UnauthorizedError{Message: message}
}

// GetStatusCode returns the HTTP status code
func (e *UnauthorizedError) GetStatusCode() int {
	return 401
}
