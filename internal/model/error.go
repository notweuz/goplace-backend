package model

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

type ServiceError struct {
	StatusCode int      `json:"status_code"`
	Message    string   `json:"message"`
	Details    []string `json:"details,omitempty"`
	Err        error    `json:"-"`
}

func (e *ServiceError) Error() string {
	return e.Message
}

func NewBadRequestError(message string, details ...string) *ServiceError {
	return &ServiceError{
		StatusCode: fiber.StatusBadRequest,
		Message:    message,
		Details:    details,
	}
}

func NewUnauthorizedError(message string) *ServiceError {
	return &ServiceError{
		StatusCode: fiber.StatusUnauthorized,
		Message:    message,
	}
}

func NewForbiddenError(message string) *ServiceError {
	return &ServiceError{
		StatusCode: fiber.StatusForbidden,
		Message:    message,
	}
}

func NewNotFoundError(message string, details ...string) *ServiceError {
	return &ServiceError{
		StatusCode: fiber.StatusNotFound,
		Message:    message,
		Details:    details,
	}
}

func NewConflictError(message string) *ServiceError {
	return &ServiceError{
		StatusCode: fiber.StatusConflict,
		Message:    message,
	}
}

func NewTooManyRequestsError(message string) *ServiceError {
	return &ServiceError{
		StatusCode: fiber.StatusTooManyRequests,
		Message:    message,
	}
}

func NewInternalServerError(message string, err error) *ServiceError {
	return &ServiceError{
		StatusCode: fiber.StatusInternalServerError,
		Message:    message,
		Err:        err,
	}
}

func NewCreatedError(statusCode int, message string, details ...string) *ServiceError {
	return &ServiceError{
		StatusCode: statusCode,
		Message:    message,
		Details:    details,
	}
}

func IsServiceError(err error) (*ServiceError, bool) {
	var se *ServiceError
	ok := errors.As(err, &se)
	return se, ok
}
