package domain

import (
	"time"

	"github.com/genefriendway/human-network-auth/internal/dto"
)

// Permission represents a permission in the IAM system.
type Permission struct {
	ID          string    `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"` // Unique ID for the permission
	PolicyID    string    `json:"policy_id" gorm:"not null"`                                 // Foreign key referencing IAMPolicy
	Resource    string    `json:"resource" gorm:"not null"`                                  // The resource this permission applies to
	Action      string    `json:"action" gorm:"not null"`                                    // The action this permission allows
	Description string    `json:"description"`                                               // Optional description of the permission
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`                          // Timestamp of permission creation
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`                          // Timestamp of last update
}

// TableName overrides the default table name for GORM.
func (m *Permission) TableName() string {
	return "iam_permissions"
}

func (m *Permission) ToDTO() dto.PermissionDTO {
	return dto.PermissionDTO{
		ID:          m.ID,
		PolicyID:    m.PolicyID,
		Resource:    m.Resource,
		Action:      m.Action,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
