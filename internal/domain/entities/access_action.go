package domain

import (
	"time"

	"gorm.io/gorm"

	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
)

// Represent a AccessAction in the IAM system.
type AccessAction struct {
	ID             string         `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name           string         `json:"name" gorm:"not null"`
	Code           string         `json:"code" gorm:"unique;not null"`
	Description    string         `json:"description"`
	OrganizationId string         `json:"organization_id" gorm:"type:uuid;not null"`
	ServiceId      string         `json:"service_id" gorm:"type:uuid;not null"`
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName overrides the default table name for GORM.
func (m *AccessAction) TableName() string {
	return "access_actions"
}

func (m *AccessAction) ToDTO() dto.AccessActionDTO {
	return dto.AccessActionDTO{
		ID: m.ID,
	}
}
