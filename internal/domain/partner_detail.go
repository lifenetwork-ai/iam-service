package domain

import (
	"time"
)

type PartnerDetail struct {
	ID          string    `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"` // UUID primary key
	AccountID   string    `json:"account_id"`
	Account     Account   `json:"account" gorm:"foreignKey:AccountID;references:ID"`
	CompanyName *string   `json:"company_name,omitempty"`
	ContactName *string   `json:"contact_name,omitempty"`
	PhoneNumber *string   `json:"phone_number,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (m *PartnerDetail) TableName() string {
	return "partner_details"
}
