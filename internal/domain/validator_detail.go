package domain

import (
	"time"
)

type ValidatorDetail struct {
	ID                     string    `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"` // UUID primary keyID                     uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	AccountID              string    `json:"account_id"`
	Account                Account   `json:"account" gorm:"foreignKey:AccountID;references:ID"`
	ValidationOrganization *string   `json:"validation_organization,omitempty"`
	ContactPerson          *string   `json:"contact_person,omitempty"`
	PhoneNumber            *string   `json:"phone_number,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

func (m *ValidatorDetail) TableName() string {
	return "validator_details"
}
