package domain

import "time"

// RefreshToken represents the structure of a refresh token in the system
type RefreshToken struct {
	ID          string    `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"` // UUID primary key
	AccountID   string    `json:"account_id"`
	HashedToken string    `json:"hashed_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (m *RefreshToken) TableName() string {
	return "refresh_tokens"
}
