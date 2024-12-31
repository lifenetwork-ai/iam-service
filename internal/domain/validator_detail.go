package domain

import (
	"time"

	"github.com/genefriendway/human-network-auth/internal/dto"
)

type ValidatorDetail struct {
	ID                     uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	AccountID              uint64    `json:"account_id"`
	ValidationOrganization string    `json:"validation_organization"`
	ContactPerson          string    `json:"contact_person"`
	PhoneNumber            string    `json:"phone_number"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

func (m *ValidatorDetail) TableName() string {
	return "validator_details"
}

func (m *ValidatorDetail) ToDTO() dto.ValidatorDetailDTO {
	return dto.ValidatorDetailDTO{
		ID:                     m.ID,
		AccountID:              m.AccountID,
		ValidationOrganization: m.ValidationOrganization,
		ContactPerson:          m.ContactPerson,
		PhoneNumber:            m.PhoneNumber,
		CreatedAt:              m.CreatedAt,
		UpdatedAt:              m.UpdatedAt,
	}
}
