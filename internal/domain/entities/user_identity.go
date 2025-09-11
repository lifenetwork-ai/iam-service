package domain

import (
	"time"
)

// Represent an identity method (email, phone, social) for a global user.
type UserIdentity struct {
	ID           string    `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	GlobalUserID string    `json:"global_user_id" gorm:"type:uuid;not null"`
	TenantID     string    `json:"tenant_id" gorm:"type:uuid;not null"`
	KratosUserID string    `json:"kratos_user_id" gorm:"type:uuid;not null"`
	Type         string    `json:"type" gorm:"type:varchar(20);not null"` // email, phone, google, wallet, etc.
	Value        string    `json:"value" gorm:"type:varchar(255);not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName overrides the default table name for GORM.
func (m *UserIdentity) TableName() string {
	return "user_identities"
}
