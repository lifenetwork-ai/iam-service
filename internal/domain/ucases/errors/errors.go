package errors

import "fmt"

type DomainError struct {
	Type    ErrorType
	Code    string
	Message string
	Details interface{}
	Cause   error
}

type ErrorType int

const (
	ErrorTypeValidation ErrorType = iota
	ErrorTypeNotFound
	ErrorTypeUnauthorized
	ErrorTypeConflict
	ErrorTypeInternal
	ErrorTypeRateLimit
)

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *DomainError) Unwrap() error {
	return e.Cause
}

// WithCause adds a cause to the error
func (e *DomainError) WithCause(cause error) *DomainError {
	e.Cause = cause
	return e
}

// WithDetails adds details to the error
func (e *DomainError) WithDetails(details interface{}) *DomainError {
	e.Details = details
	return e
}

// NewValidationError creates a validation error
func NewValidationError(code, message string, details interface{}) *DomainError {
	return &DomainError{
		Type:    ErrorTypeValidation,
		Code:    code,
		Message: message,
		Details: details,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(code, resource string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeNotFound,
		Code:    code,
		Message: fmt.Sprintf("%s not found", resource),
	}
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(code, message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeUnauthorized,
		Code:    code,
		Message: message,
	}
}

// NewConflictError creates a conflict error
func NewConflictError(code, message string, details interface{}) *DomainError {
	return &DomainError{
		Type:    ErrorTypeConflict,
		Code:    code,
		Message: message,
		Details: details,
	}
}

// NewInternalError creates an internal error
func NewInternalError(code, message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeInternal,
		Code:    code,
		Message: message,
	}
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(code, message string, details interface{}) *DomainError {
	return &DomainError{
		Type:    ErrorTypeRateLimit,
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Wrap wraps an error with a domain error
func Wrap(err error, errorType ErrorType, code, message string) *DomainError {
	return &DomainError{
		Type:    errorType,
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// WrapInternal wraps an error as an internal error
func WrapInternal(err error, code, message string) *DomainError {
	return Wrap(err, ErrorTypeInternal, code, message)
}

// WrapValidation wraps an error as a validation error
func WrapValidation(err error, code, message string, details interface{}) *DomainError {
	return &DomainError{
		Type:    ErrorTypeValidation,
		Code:    code,
		Message: message,
		Details: details,
		Cause:   err,
	}
}
