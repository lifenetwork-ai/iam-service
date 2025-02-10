package domain

import (
	"time"

	"github.com/genefriendway/human-network-iam/internal/dto"
)

// Organization represents a Organization in the IAM system.
type Organization struct {
	ID          string    `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"` // Unique ID for the Organization
	Name        string    `json:"name" gorm:"not null"`                                      // Organization's name
	Code        string    `json:"code" gorm:"not null"`                                      // Organization's code
	Description string    `json:"description"`                                               // Optional description of the Organization
	ParentID    string    `json:"parent_id" gorm:"type:uuid"`                                // Parent Organization ID
	ParentPath  string    `json:"parent_path"`                                               // Parent Organization path
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`                          // Timestamp of Organization creation
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`                          // Timestamp of last update
}

// TableName overrides the default table name for GORM.
func (m *Organization) TableName() string {
	return "identity_organizations"
}

func (m *Organization) ToDTO() dto.OrganizationDTO {
	return dto.OrganizationDTO{
		ID:          m.ID,
		Name:        m.Name,
		Code:        m.Code,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
