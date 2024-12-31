package dto

import "time"

type AccountDTO struct {
	ID            uint64    `json:"id"`
	Email         string    `json:"email"`
	Role          string    `json:"role"`                     // USER, PARTNER, CUSTOMER, VALIDATOR
	APIKey        *string   `json:"api_key,omitempty"`        // Nullable
	OAuthProvider *string   `json:"oauth_provider,omitempty"` // Nullable
	OAuthID       *string   `json:"oauth_id,omitempty"`       // Nullable
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type AccountDetailDTO struct {
	ID                     uint64     `json:"id"`
	AccountID              uint64     `json:"account_id"`
	Role                   string     `json:"role"`
	ValidationOrganization *string    `json:"validation_organization,omitempty"`
	CompanyName            *string    `json:"company_name,omitempty"`
	ContactName            *string    `json:"contact_name,omitempty"`
	FirstName              *string    `json:"first_name,omitempty"`
	LastName               *string    `json:"last_name,omitempty"`
	DateOfBirth            *time.Time `json:"date_of_birth,omitempty"`
	PhoneNumber            *string    `json:"phone_number,omitempty"`
	Industry               *string    `json:"industry,omitempty"`
	OrganizationName       *string    `json:"organization_name,omitempty"`
}

type RegisterAccountDTO struct {
	Email                  string     `json:"email" validate:"required,email"`
	Password               string     `json:"password" validate:"required"`
	Role                   string     `json:"role" validate:"required"`
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
