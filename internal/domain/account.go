package domain

import (
	"time"
)

type Account struct {
	ID            uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	Email         string    `json:"email"`
	PasswordHash  *string   `json:"password_hash,omitempty"`         // Nullable for OAuth or API Key accounts
	APIKey        *string   `json:"api_key,omitempty" gorm:"unique"` // Nullable, used for API-based roles
	Role          string    `json:"role"`
	OAuthProvider *string   `json:"oauth_provider,omitempty"` // Nullable, stores OAuth provider name (e.g., Google, Facebook)
	OAuthID       *string   `json:"oauth_id,omitempty"`       // Nullable, stores ID from OAuth provider
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (a *Account) TableName() string {
	return "accounts"
}
