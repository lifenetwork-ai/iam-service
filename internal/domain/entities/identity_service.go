package domain

import (
	"time"

	"gorm.io/gorm"

	"github.com/genefriendway/human-network-iam/internal/delivery/dto"
)

// Represent a IdentityService in the IAM system.
type IdentityService struct {
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
func (m *IdentityService) TableName() string {
	return "identity_services"
}

func (m *IdentityService) ToDTO() dto.IdentityServiceDTO {
	return dto.IdentityServiceDTO{
		ID: m.ID,
	}
}
