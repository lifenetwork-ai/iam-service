package domain

import (
	"time"
)

// Represent a tenant-specific user identifier mapping to global user.
type UserIdentifierMapping struct {
	ID           string    `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	GlobalUserID string    `json:"global_user_id" gorm:"type:uuid;not null"`
	Tenant       string    `json:"tenant" gorm:"type:varchar(25);not null"`
	TenantUserID string    `json:"tenant_user_id" gorm:"type:varchar(36);not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName overrides the default table name for GORM.
func (m *UserIdentifierMapping) TableName() string {
	return "user_identifier_mapping"
}
