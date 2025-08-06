package domain

import "time"

type ZaloToken struct {
	ID           uint   `gorm:"primaryKey"`
	AccessToken  string `gorm:"type:text;not null"`
	RefreshToken string `gorm:"type:text;not null"`
	ExpiresAt    time.Time
	UpdatedAt    time.Time
}

func (ZaloToken) TableName() string {
	return "zalo_tokens"
}
