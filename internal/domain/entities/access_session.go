package domain

import (
	"time"

	"gorm.io/gorm"

	"github.com/genefriendway/human-network-iam/internal/delivery/dto"
)

// Represent a AccessSession in the IAM system.
type AccessSession struct {
	ID               string         `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserId           string         `json:"user_id" gorm:"type:uuid;not null"`
	OrganizationId   string         `json:"organization_id" gorm:"type:uuid;not null"`
	AccessToken      string         `json:"access_token" gorm:"type:uuid;not null"`
	RefreshToken     string         `json:"refresh_token" gorm:"type:uuid;not null"`
	AccessExpiredAt  time.Time      `json:"access_expired_at" gorm:"not null"`
	RefreshExpiredAt time.Time      `json:"refresh_expired_at" gorm:"not null"`
	LastRevokedAt    time.Time      `json:"last_revoked_at" gorm:"not null"`
	DeviceId         string         `json:"device_id" gorm:"type:uuid;not null"`
	FirebaseToken    string         `json:"firebase_token" gorm:"type:text"`
	CreatedAt        time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `json:"deleted_at" gorm:"index"`
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
