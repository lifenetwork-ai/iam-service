package dto

import "time"

// LoginPayloadDTO defines the payload for the login request
type LoginPayloadDTO struct {
	Email    string `json:"email" validate:"required,email"` // User email
	Password string `json:"password" validate:"required"`    // User password
}

// RegisterPayloadtDTO defines the payload for the register request
type RegisterPayloadDTO struct {
	Email                  string     `json:"email" validate:"required,email"`
	Password               string     `json:"password" validate:"required"`
	Role                   string     `json:"role" validate:"required"` // USER, PARTNER, CUSTOMER, VALIDATOR
	FirstName              *string    `json:"first_name,omitempty"`
	LastName               *string    `json:"last_name,omitempty"`
	DateOfBirth            *time.Time `json:"date_of_birth,omitempty"`
	PhoneNumber            *string    `json:"phone_number,omitempty"`
	CompanyName            *string    `json:"company_name,omitempty"`
	ContactName            *string    `json:"contact_name,omitempty"`
	OrganizationName       *string    `json:"organization_name,omitempty"`
	Industry               *string    `json:"industry,omitempty"`
	ValidationOrganization *string    `json:"validation_organization,omitempty"`
}

// RefreshTokenPayloadDTO defines the payload for refreshing tokens request
type RefreshTokenPayloadDTO struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutPayloadDTO defines the payload for the logout request
type LogoutPayloadDTO struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
