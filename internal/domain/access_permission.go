package domain

import (
	"time"

	"gorm.io/gorm"

	"github.com/genefriendway/human-network-iam/internal/dto"
)

// Represent a AccessPermission in the IAM system.
type AccessPermission struct {
	ID             string         `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name           string         `json:"name" gorm:"not null"`
	Code           string         `json:"code" gorm:"unique;not null"`
	Description    string         `json:"description"`
	OrganizationId string         `json:"organization_id" gorm:"type:uuid;not null"`
	ServiceId      string         `json:"service_id" gorm:"type:uuid;not null"`
	Policies       []AccessPolicy `json:"policies" gorm:"foreignKey:PermissionId"`
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName overrides the default table name for GORM.
func (m *AccessPermission) TableName() string {
	return "access_permissions"
}

func (m *AccessPermission) ToDTO() dto.AccessPermissionDTO {
	return dto.AccessPermissionDTO{
		ID: m.ID,
	}
}
