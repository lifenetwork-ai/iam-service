package domain

import "time"

// AccountPolicy represents an assignment of a policy to an account.
type AccountPolicy struct {
	ID        string    `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"` // Unique ID for the account-policy assignment
	AccountID string    `json:"account_id" gorm:"not null"`                                // Foreign key referencing the Account
	PolicyID  string    `json:"policy_id" gorm:"not null"`                                 // Foreign key referencing the IAMPolicy
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`                          // Timestamp of assignment creation
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`                          // Timestamp of last update
}

// TableName overrides the default table name for GORM.
func (m *AccountPolicy) TableName() string {
	return "account_policies"
}
