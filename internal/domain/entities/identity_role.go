package domain

import (
	"time"

	"gorm.io/gorm"

	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
)

// Represent a IdentityRole in the IAM system.
type IdentityRole struct {
	ID             string             `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name           string             `json:"name" gorm:"not null"`
	Code           string             `json:"code" gorm:"unique;not null"`
	Description    string             `json:"description"`
	OrganizationId string             `json:"organization_id" gorm:"type:uuid;not null"`
	Permissions    []AccessPermission `json:"permissions" gorm:"foreignKey:RoleId"`
	Policies       []AccessPolicy     `json:"policies" gorm:"foreignKey:RoleId"`
	CreatedAt      time.Time          `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time          `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt     `json:"deleted_at" gorm:"index"`
}

// TableName overrides the default table name for GORM.
func (m *IdentityRole) TableName() string {
	return "identity_roles"
}

func (m *IdentityRole) ToDTO() dto.IdentityRoleDTO {
	return dto.IdentityRoleDTO{
		ID: m.ID,
	}
}
