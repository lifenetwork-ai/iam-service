package domain

import (
	"time"
)

type DataUtilizer struct {
	ID               string    `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"` // UUID primary key
	AccountID        string    `json:"account_id"`
	Account          Account   `json:"account" gorm:"foreignKey:AccountID;references:ID"`
	OrganizationName *string   `json:"organization_name,omitempty"`
	Industry         *string   `json:"industry,omitempty"`
	ContactName      *string   `json:"contact_name,omitempty"`
	PhoneNumber      *string   `json:"phone_number,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (m *DataUtilizer) TableName() string {
	return "data_utilizers"
}
