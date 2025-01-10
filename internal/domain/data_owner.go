package domain

import (
	"time"
)

type DataOwner struct {
	ID          string    `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"` // UUID primary key
	AccountID   string    `json:"account_id"`
	Account     Account   `json:"account" gorm:"foreignKey:AccountID;references:ID"`
	FirstName   *string   `json:"first_name,omitempty"`
	LastName    *string   `json:"last_name,omitempty"`
	PhoneNumber *string   `json:"phone_number,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (m *DataOwner) TableName() string {
	return "data_owners"
}
