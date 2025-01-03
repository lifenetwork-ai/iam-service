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

// UpdateRolePayloadDTO defines the payload for updating the role of an account
type UpdateRolePayloadDTO struct {
	Role        string                `json:"role" validate:"required"` // USER, PARTNER, CUSTOMER, VALIDATOR
	RoleDetails RoleDetailsPayloadDTO `json:"role_details,omitempty"`   // Role-specific details
}

// RoleDetailsPayloadDTO defines the payload for the role-specific details
type RoleDetailsPayloadDTO struct {
	// Common fields
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`

	// Partner fields
	CompanyName string `json:"company_name,omitempty"`
	ContactName string `json:"contact_name,omitempty"`

	// Customer fields
	OrganizationName string `json:"organization_name,omitempty"`
	Industry         string `json:"industry,omitempty"`

	// Validator fields
	ValidationOrganization string `json:"validation_organization,omitempty"`
}
