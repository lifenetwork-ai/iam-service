package domain

import (
	"time"

	"github.com/google/uuid"
)

// AdminAccount represents an admin account in the system
type AdminAccount struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Username     string    `gorm:"type:varchar(255);not null;unique"`
	PasswordHash string    `gorm:"type:varchar(255);not null"`
	Role         string    `gorm:"type:varchar(50);not null;default:'ADMIN'"`
	Status       string    `gorm:"type:varchar(50);not null;default:'active'"`
	CreatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}
