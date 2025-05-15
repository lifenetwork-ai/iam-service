package domain

import (
	"time"

	"gorm.io/gorm"

	"github.com/genefriendway/human-network-iam/internal/delivery/dto"
)

// Represent a IdentityUser in the IAM system.
type IdentityUser struct {
	ID             string    `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Seed           string    `json:"seed"`
	OrganizationId string    `json:"organization_id" gorm:"type:uuid;not null"`
	UserName       string    `json:"user_name" gorm:"not null"`
	Email          string    `json:"email" gorm:"not null"`
	Phone          string    `json:"phone" gorm:"not null"`
	PasswordHash   string    `json:"password_hash" gorm:"not null"`
	Status         bool      `json:"status" gorm:"not null;default:true"`
	Name           string    `json:"name" gorm:"not null"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	FullName       string    `json:"full_name"`
	LastLoginAt    time.Time `json:"last_login_at"`

	// External ID
	SelfAuthenticateID string `json:"self_authenticate_id" gorm:"column:self_authenticate_id"`
	GoogleID           string `json:"google_id"`
	FacebookID         string `json:"facebook_id"`
	AppleID            string `json:"apple_id"`

	// Audit fields
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName overrides the default table name for GORM.
func (m *IdentityUser) TableName() string {
	return "identity_users"
}

func (m *IdentityUser) ToDTO() dto.IdentityUserDTO {
	return dto.IdentityUserDTO{
		ID:        m.ID,
		Seed:      m.Seed,
		Name:      m.Name,
		UserName:  m.UserName,
		Email:     m.Email,
		Phone:     m.Phone,
		Status:    m.Status,
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
	}
}
