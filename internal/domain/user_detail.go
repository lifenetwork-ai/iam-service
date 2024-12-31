package domain

import (
	"time"

	"github.com/genefriendway/human-network-auth/internal/dto"
)

type UserDetail struct {
	ID          uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	AccountID   uint64    `json:"account_id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DateOfBirth time.Time `json:"date_of_birth"`
	PhoneNumber string    `json:"phone_number"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (m *UserDetail) TableName() string {
	return "user_details"
}

func (m *UserDetail) ToDTO() dto.UserDetailDTO {
	return dto.UserDetailDTO{
		ID:          m.ID,
		AccountID:   m.AccountID,
		FirstName:   m.FirstName,
		LastName:    m.LastName,
		DateOfBirth: m.DateOfBirth,
		PhoneNumber: m.PhoneNumber,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
