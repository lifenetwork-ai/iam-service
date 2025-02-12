package domain

import (
	"time"

	"github.com/genefriendway/human-network-iam/internal/dto"
	"gorm.io/gorm"
)

// Represent a AccessPolicy in the IAM system.
type AccessPolicy struct {
	ID          string         `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName overrides the default table name for GORM.
func (m *AccessPolicy) TableName() string {
	return "access_policies"
}

func (m *AccessPolicy) ToDTO() dto.AccessPolicyDTO {
	return dto.AccessPolicyDTO{
		ID:          m.ID,
	}
}
