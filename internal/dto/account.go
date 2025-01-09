package dto

type AccountDTO struct {
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	Username      string  `json:"username"`
	Role          string  `json:"role"`                     // USER, PARTNER, CUSTOMER, VALIDATOR
	PublicKey     *string `json:"public_key,omitempty"`     // Nullable
	PrivateKey    *string `json:"private_key,omitempty"`    // Nullable
	APIKey        *string `json:"api_key,omitempty"`        // Nullable
	OAuthProvider *string `json:"oauth_provider,omitempty"` // Nullable
	OAuthID       *string `json:"oauth_id,omitempty"`       // Nullable
}

type AccountDetailDTO struct {
	ID                     *string    `json:"id,omitempty"`
	Account                AccountDTO `json:"account"`
	ValidationOrganization *string    `json:"validation_organization,omitempty"`
	CompanyName            *string    `json:"company_name,omitempty"`
	ContactName            *string    `json:"contact_name,omitempty"`
	FirstName              *string    `json:"first_name,omitempty"`
	LastName               *string    `json:"last_name,omitempty"`
	PhoneNumber            *string    `json:"phone_number,omitempty"`
	Industry               *string    `json:"industry,omitempty"`
	OrganizationName       *string    `json:"organization_name,omitempty"`
}
