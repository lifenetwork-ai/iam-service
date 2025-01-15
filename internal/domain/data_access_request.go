package domain

import (
	"time"

	"github.com/genefriendway/human-network-auth/internal/dto"
)

type DataAccessRequest struct {
	ID                 string                       `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	RequestAccountID   string                       `json:"request_account_id" gorm:"not null"`               // Account whose data is being requested
	FileID             string                       `json:"file_id" gorm:"type:uuid;not null"`                // ID of the file being accessed
	ReasonForRequest   string                       `json:"reason_for_request" gorm:"not null"`               // Reason for the request
	Status             string                       `json:"status" gorm:"type:varchar(20);default:'PENDING'"` // Request status (PENDING, APPROVED, REJECTED)
	ReasonForRejection *string                      `json:"reason_for_rejection,omitempty"`                   // Reason for rejection (optional)
	Requesters         []DataAccessRequestRequester `json:"requesters" gorm:"foreignKey:RequestID"`           // Linked requesters
	CreatedAt          time.Time                    `json:"created_at" gorm:"autoCreateTime"`                 // Automatically set creation timestamp
	UpdatedAt          time.Time                    `json:"updated_at" gorm:"autoUpdateTime"`                 // Automatically update timestamp on changes
}

// TableName overrides the default table name for GORM
func (m *DataAccessRequest) TableName() string {
	return "data_access_requests"
}

// ToDTO converts a DataAccessRequest domain model to a DataAccessRequestDTO.
func (m *DataAccessRequest) ToDTO() *dto.DataAccessRequestDTO {
	requesters := make([]dto.AccountDTO, len(m.Requesters))
	for i, requester := range m.Requesters {
		requesters[i] = *requester.RequesterAccount.ToDTO()
	}

	return &dto.DataAccessRequestDTO{
		ID:                 m.ID,
		RequestAccountID:   m.RequestAccountID,
		FileID:             m.FileID,
		Requesters:         requesters,
		ReasonForRequest:   m.ReasonForRequest,
		Status:             m.Status,
		ReasonForRejection: m.ReasonForRejection,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
}
