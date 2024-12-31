package domain

import "time"

// RefreshToken represents the structure of a refresh token in the system
type RefreshToken struct {
	ID          uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	AccountID   uint64    `json:"account_id"`
	HashedToken string    `json:"hashed_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
