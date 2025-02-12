package domain

import (
	"time"

	"github.com/genefriendway/human-network-iam/internal/dto"
	"gorm.io/gorm"
)

// Represent a AccessAction in the IAM system.
type AccessAction struct {
	ID          string         `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName overrides the default table name for GORM.
func (m *AccessAction) TableName() string {
	return "access_actions"
}

func (m *AccessAction) ToDTO() dto.AccessActionDTO {
	return dto.AccessActionDTO{
		ID:          m.ID,
	}
}
