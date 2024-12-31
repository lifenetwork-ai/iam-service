package dto

import "time"

type ValidatorDetailDTO struct {
	ID                     uint64    `json:"id"`
	AccountID              uint64    `json:"account_id"`
	ValidationOrganization string    `json:"validation_organization"`
	ContactPerson          string    `json:"contact_person"`
	PhoneNumber            string    `json:"phone_number"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}
