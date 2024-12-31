package domain

import (
	"time"

	"github.com/genefriendway/human-network-auth/internal/dto"
)

type CustomerDetail struct {
	ID               uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	AccountID        uint64    `json:"account_id"`
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

func (m *CustomerDetail) ToDTO() dto.CustomerDetailDTO {
	return dto.CustomerDetailDTO{
		ID:               m.ID,
		AccountID:        m.AccountID,
		OrganizationName: m.OrganizationName,
		Industry:         m.Industry,
		ContactName:      m.ContactName,
		PhoneNumber:      m.PhoneNumber,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}
