package domain

import (
	"time"

	"github.com/genefriendway/human-network-iam/internal/dto"
	"gorm.io/gorm"
)

// Represent a IdentityRole in the IAM system.
type IdentityRole struct {
	ID          string         `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName overrides the default table name for GORM.
func (m *IdentityRole) TableName() string {
	return "identity_roles"
}

func (m *IdentityRole) ToDTO() dto.IdentityRoleDTO {
	return dto.IdentityRoleDTO{
		ID:          m.ID,
	}
}
