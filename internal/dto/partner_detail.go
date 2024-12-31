package dto

import "time"

type PartnerDetailDTO struct {
	ID          uint64    `json:"id"`
	AccountID   uint64    `json:"account_id"`
	CompanyName string    `json:"company_name"`
	ContactName string    `json:"contact_name"`
	PhoneNumber string    `json:"phone_number"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
