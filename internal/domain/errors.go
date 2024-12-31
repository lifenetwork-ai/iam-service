package domain

import "errors"

// ErrInvalidCredentials represents an invalid login attempt
var ErrInvalidCredentials = errors.New("invalid credentials")
