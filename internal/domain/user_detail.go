package domain

import (
	"time"
)

type UserDetail struct {
	ID          uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	AccountID   uint64    `json:"account_id"`
	Account     Account   `json:"account" gorm:"foreignKey:AccountID;references:ID"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	PhoneNumber string    `json:"phone_number"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (m *UserDetail) TableName() string {
	return "user_details"
}
