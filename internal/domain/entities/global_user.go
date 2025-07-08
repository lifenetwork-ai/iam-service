package domain

import (
	"time"
)

// Represent a global user in the IAM system.
type GlobalUser struct {
	ID         string         `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Identities []UserIdentity `json:"identities" gorm:"foreignKey:GlobalUserID;references:ID"`
	CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName overrides the default table name for GORM.
func (m *GlobalUser) TableName() string {
	return "global_users"
}
