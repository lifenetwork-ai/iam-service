package domain

import (
	"time"
)

type Validator struct {
	ID                     *string   `json:"id,omitempty" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"` // UUID primary keyID                     uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	AccountID              string    `json:"account_id"`
	Account                Account   `json:"account" gorm:"foreignKey:AccountID;references:ID"`
	ValidationOrganization *string   `json:"validation_organization,omitempty"`
	ContactPerson          *string   `json:"contact_person,omitempty"`
	PhoneNumber            *string   `json:"phone_number,omitempty"`
	IsActive               bool      `json:"is_active" gorm:"default:true"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

func (m *Validator) TableName() string {
	return "validators"
}
