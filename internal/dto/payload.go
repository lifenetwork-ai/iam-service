package dto

import (
	"github.com/genefriendway/human-network-auth/constants"
)

// LoginPayloadDTO defines the payload for the login request
type LoginPayloadDTO struct {
	Identifier     string                   `json:"identifier" validate:"required"`      // Identifier (email, username, or phone number)
	Password       string                   `json:"password" validate:"required"`        // User password
	IdentifierType constants.IdentifierType `json:"identifier_type" validate:"required"` // Type of identifier: "email", "username", or "phone"
}

// RegisterPayloadtDTO defines the payload for the register request
type RegisterPayloadDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RefreshTokenPayloadDTO defines the payload for refreshing tokens request
type RefreshTokenPayloadDTO struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutPayloadDTO defines the payload for the logout request
type LogoutPayloadDTO struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
