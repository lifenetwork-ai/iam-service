package domain

import "errors"

// General Errors
var (
	ErrInvalidCredentials      = errors.New("invalid credentials")      // Login failure
	ErrInsufficientPermissions = errors.New("insufficient permissions") // Authorization failure
)

// Token Errors
var (
	ErrInvalidToken  = errors.New("invalid token")   // Token format or content is invalid
	ErrExpiredToken  = errors.New("expired token")   // Token is no longer valid
	ErrTokenNotFound = errors.New("token not found") // Token missing in the request
)

// Data Errors
var (
	ErrDataNotFound      = errors.New("data not found")        // Resource not found
	ErrAlreadyExists     = errors.New("record already exists") // Duplicate entry
	ErrInvalidParameters = errors.New("invalid parameters")    // Validation error
)
