package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Represent a tenant-specific user identifier mapping to global user.
type UserIdentifierMapping struct {
	ID           string    `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	GlobalUserID string    `json:"global_user_id" gorm:"type:uuid;not null"`
	Lang         string    `json:"lang" gorm:"type:varchar(10);not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate is a GORM hook that generates a UUID for the UserIdentifierMapping if it is not set.
// Needed for SQLite tests.
func (u *UserIdentifierMapping) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		uuid, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		u.ID = uuid.String()
	}
	return
}

// TableName overrides the default table name for GORM.
func (m *UserIdentifierMapping) TableName() string {
	return "user_identifier_mapping"
}
