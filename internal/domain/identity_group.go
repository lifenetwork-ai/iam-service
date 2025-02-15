package domain

import (
	"time"

	"github.com/genefriendway/human-network-iam/internal/dto"
	"gorm.io/gorm"
)

// Represent a IdentityGroup in the IAM system.
type IdentityGroup struct {
	ID             string         `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name           string         `json:"name" gorm:"not null"`
	Code           string         `json:"code" gorm:"unique;not null"`
	Description    string         `json:"description"`
	OrganizationId string         `json:"organization_id" gorm:"type:uuid;not null"`
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName overrides the default table name for GORM.
func (m *IdentityGroup) TableName() string {
	return "identity_groups"
}

func (m *IdentityGroup) ToDTO() dto.IdentityGroupDTO {
	return dto.IdentityGroupDTO{
		ID: m.ID,
	}
}
