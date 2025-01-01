package domain

import (
	"time"
)

type CustomerDetail struct {
	ID               uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	AccountID        uint64    `json:"account_id"`
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
