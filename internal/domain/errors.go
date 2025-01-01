package domain

import "errors"

// ErrInvalidCredentials represents an invalid login attempt
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrAccountAlreadyExists is returned when an account with the same email already exists
var ErrAccountAlreadyExists = errors.New("account already exists")
