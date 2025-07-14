package domain

import (
	"time"
)

// Represent a change log for user identity updates.
type UserIdentityChangeLog struct {
	ID           string    `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	GlobalUserID string    `json:"global_user_id" gorm:"type:uuid;not null"`
	TenantID     string    `json:"tenant_id" gorm:"type:uuid;not null"`
	IdentityType string    `json:"identity_type" gorm:"type:varchar(20);not null"`
	OldValue     string    `json:"old_value" gorm:"type:varchar(255)"`
	NewValue     string    `json:"new_value" gorm:"type:varchar(255);not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName overrides the default table name for GORM.
func (m *UserIdentityChangeLog) TableName() string {
	return "user_identity_change_logs"
}
