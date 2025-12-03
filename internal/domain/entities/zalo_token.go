package domain

import (
	"time"

	"github.com/google/uuid"
)

type ZaloToken struct {
	ID            uint      `gorm:"primaryKey"`
	TenantID      uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	AppID         string    `gorm:"type:varchar(255);not null"`
	SecretKey     string    `gorm:"type:text;not null"`
	AccessToken   string    `gorm:"type:text;not null"`
	RefreshToken  string    `gorm:"type:text;not null"`
	OtpTemplateID string    `gorm:"type:varchar(255)"`
	ExpiresAt     time.Time `gorm:"not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (ZaloToken) TableName() string {
	return "zalo_tokens"
}
