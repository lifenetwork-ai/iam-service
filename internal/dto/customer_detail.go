package dto

import "time"

type CustomerDetailDTO struct {
	ID               uint64    `json:"id"`
	AccountID        uint64    `json:"account_id"`
	OrganizationName string    `json:"organization_name"`
	Industry         string    `json:"industry"`
	ContactName      string    `json:"contact_name"`
	PhoneNumber      string    `json:"phone_number"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
