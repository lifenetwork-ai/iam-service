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
