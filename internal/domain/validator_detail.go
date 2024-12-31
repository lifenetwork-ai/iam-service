package domain

import "time"

type ValidatorDetail struct {
	ID                     uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	AccountID              uint64    `json:"account_id"`
	ValidationOrganization string    `json:"validation_organization"`
	ContactPerson          string    `json:"contact_person"`
	PhoneNumber            string    `json:"phone_number"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

func (v *ValidatorDetail) TableName() string {
	return "validator_details"
}
