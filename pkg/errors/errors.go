package errors

import "errors"

type AppError struct {
	Message string
	Code    string // e.g., "NOT_FOUND", "VALIDATION_ERROR", "CONFLICT"
}

func (e *AppError) Error() string {
	return e.Message
}

func NewNotFoundError(message string) error {
	return &AppError{
		Message: message,
		Code:    "NOT_FOUND",
	}
}

func IsNotFoundError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == "NOT_FOUND"
}

func NewValidationError(message string) error {
	return &AppError{
		Message: message,
		Code:    "VALIDATION_ERROR",
	}
}

func IsValidationError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == "VALIDATION_ERROR"
}

func NewConflictError(message string) error {
	return &AppError{
		Message: message,
		Code:    "CONFLICT",
	}
}

func IsConflictError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == "CONFLICT"
}
