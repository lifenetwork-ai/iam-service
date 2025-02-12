package domain

import (
	"time"

	"github.com/genefriendway/human-network-iam/internal/dto"
	"gorm.io/gorm"
)

// Represent a AccessPermission in the IAM system.
type AccessPermission struct {
	ID          string         `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName overrides the default table name for GORM.
func (m *AccessPermission) TableName() string {
	return "access_permissions"
}

func (m *AccessPermission) ToDTO() dto.AccessPermissionDTO {
	return dto.AccessPermissionDTO{
		ID:          m.ID,
	}
}
