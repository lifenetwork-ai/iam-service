package domain

import "errors"

// ErrInvalidCredentials represents an invalid login attempt
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrAccountAlreadyExists is returned when an account with the same email already exists
var ErrAccountAlreadyExists = errors.New("account already exists")

// ErrInvalidToken is returned when a provided token is invalid.
var ErrInvalidToken = errors.New("invalid token")

// ErrExpiredToken is returned when a provided token has expired.
var ErrExpiredToken = errors.New("expired token")

// ErrTokenNotFound is returned when a token is not found.
var ErrTokenNotFound = errors.New("token not found")

// ErrInsufficientPermissions is returned when a user does not have the required permissions.
var ErrInsufficientPermissions = errors.New("insufficient permissions")
