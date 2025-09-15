package domain

import (
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
)

// Represent a global user in the IAM system.
type GlobalUser struct {
	ID         string         `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Identities []UserIdentity `json:"identities" gorm:"foreignKey:GlobalUserID;references:ID"`
	CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

func (u *GlobalUser) BeforeCreate(tx *gorm.DB) (err error) {
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
func (m *GlobalUser) TableName() string {
	return "global_users"
}
