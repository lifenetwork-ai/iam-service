package dto

import "time"

// DataAccessRequestDTO represents the structure for returning data access request information.
type DataAccessRequestDTO struct {
	ID                 string       `json:"id"`                             // Unique identifier for the request
	RequestAccountID   string       `json:"request_account_id"`             // ID of the account being accessed
	FileID             string       `json:"file_id"`                        // ID of the file being accessed
	Requesters         []AccountDTO `json:"requesters"`                     // List of accounts making the request
	ReasonForRequest   string       `json:"reason_for_request"`             // Reason for the request
	Status             string       `json:"status"`                         // Status of the request (PENDING, APPROVED, REJECTED)
	ReasonForRejection *string      `json:"reason_for_rejection,omitempty"` // Optional reason for rejection
	CreatedAt          time.Time    `json:"created_at"`                     // Timestamp of when the request was created
	UpdatedAt          time.Time    `json:"updated_at"`                     // Timestamp of when the request was last updated
}
