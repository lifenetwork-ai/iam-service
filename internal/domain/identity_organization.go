package domain

import (
	"time"

	"gorm.io/gorm"

	"github.com/genefriendway/human-network-iam/internal/dto"
)

// Represents a IdentityOrganization in the IAM system.
type IdentityOrganization struct {
	ID          string                `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"` // Unique ID for the IdentityOrganization
	Name        string                `json:"name" gorm:"not null"`                                      // IdentityOrganization's name
	Code        string                `json:"code" gorm:"not null"`                                      // IdentityOrganization's code
	Description string                `json:"description"`                                               // Optional description of the IdentityOrganization
	Parent      *IdentityOrganization `json:"parent,omitempty" gorm:"foreignKey:ParentID"`               // Parent IdentityOrganization
	ParentID    string                `json:"parent_id"`                                                 // Parent IdentityOrganization ID
	ParentPath  string                `json:"parent_path"`                                               // Parent IdentityOrganization path
	CreatedAt   time.Time             `json:"created_at" gorm:"autoCreateTime"`                          // Timestamp of IdentityOrganization creation
	UpdatedAt   time.Time             `json:"updated_at" gorm:"autoUpdateTime"`                          // Timestamp of last update
	DeletedAt   gorm.DeletedAt        `json:"deleted_at" gorm:"index"`                                   // Timestamp of deletion
}

// TableName overrides the default table name for GORM.
func (m *IdentityOrganization) TableName() string {
	return "identity_organizations"
}

func (m *IdentityOrganization) ToDTO() dto.IdentityOrganizationDTO {
	return dto.IdentityOrganizationDTO{
		ID:          m.ID,
		Name:        m.Name,
		Code:        m.Code,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
