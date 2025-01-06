package domain

import "time"

type DataAccessRequest struct {
	ID                 string    `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	RequestAccountID   string    `json:"request_account_id" gorm:"not null"`               // Account whose data is being requested
	RequesterAccountID string    `json:"requester_account_id" gorm:"not null"`             // Account making the request
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
