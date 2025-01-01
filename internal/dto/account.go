package dto

import "time"

type AccountDTO struct {
	ID            uint64  `json:"id"`
	Email         string  `json:"email"`
	Role          string  `json:"role"`                     // USER, PARTNER, CUSTOMER, VALIDATOR
	APIKey        *string `json:"api_key,omitempty"`        // Nullable
	OAuthProvider *string `json:"oauth_provider,omitempty"` // Nullable
	OAuthID       *string `json:"oauth_id,omitempty"`       // Nullable
}

type AccountDetailDTO struct {
	ID                     uint64     `json:"id"`
	Account                AccountDTO `json:"account"`
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
