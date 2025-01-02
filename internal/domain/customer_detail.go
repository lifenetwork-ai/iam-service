package domain

import (
	"time"
)

type CustomerDetail struct {
	ID               string    `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"` // UUID primary key
	AccountID        string    `json:"account_id"`
	Account          Account   `json:"account" gorm:"foreignKey:AccountID;references:ID"`
	OrganizationName string    `json:"organization_name"`
	Industry         string    `json:"industry"`
	ContactName      string    `json:"contact_name"`
	PhoneNumber      string    `json:"phone_number"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (m *CustomerDetail) TableName() string {
	return "customer_details"
}
