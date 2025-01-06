package domain

import (
	"time"

	"github.com/genefriendway/human-network-auth/internal/dto"
)

type DataAccessRequest struct {
	ID                 string    `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	RequestAccountID   string    `json:"request_account_id" gorm:"not null"`   // Account whose data is being requested
	RequesterAccountID string    `json:"requester_account_id" gorm:"not null"` // ID of the requester account
	RequesterAccount   Account   `json:"requester_account" gorm:"foreignKey:RequesterAccountID;references:ID"`
	RequesterRole      string    `json:"requester_role" gorm:"type:varchar(20);not null"`  // Role of the requester
	ReasonForRequest   string    `json:"reason_for_request" gorm:"not null"`               // Reason for the request
	Status             string    `json:"status" gorm:"type:varchar(20);default:'PENDING'"` // Request status (PENDING, APPROVED, REJECTED)
	ReasonForRejection *string   `json:"reason_for_rejection,omitempty"`                   // Reason for rejection (optional)
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime"`                 // Automatically set creation timestamp
	UpdatedAt          time.Time `json:"updated_at" gorm:"autoUpdateTime"`                 // Automatically update timestamp on changes
}

// TableName overrides the default table name for GORM
func (m *DataAccessRequest) TableName() string {
	return "data_access_requests"
}

// ToDTO converts a DataAccessRequest domain model to a DataAccessRequestDTO.
func (m *DataAccessRequest) ToDTO() *dto.DataAccessRequestDTO {
	return &dto.DataAccessRequestDTO{
		RequestAccountID:   m.RequestAccountID,
		RequesterAccount:   *m.RequesterAccount.ToDTO(),
		ReasonForRequest:   m.ReasonForRequest,
		Status:             m.Status,
		ReasonForRejection: m.ReasonForRejection,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
}
