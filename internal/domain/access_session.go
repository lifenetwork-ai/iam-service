package domain

import (
	"time"

	"gorm.io/gorm"

	"github.com/genefriendway/human-network-iam/internal/dto"
)

// Represent a AccessSession in the IAM system.
type AccessSession struct {
	ID        string         `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName overrides the default table name for GORM.
func (m *AccessSession) TableName() string {
	return "access_sessions"
}

func (m *AccessSession) ToDTO() dto.AccessSessionDTO {
	return dto.AccessSessionDTO{
		ID: m.ID,
	}
}
