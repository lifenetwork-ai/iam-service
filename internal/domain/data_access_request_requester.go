package domain

type DataAccessRequestRequester struct {
	ID                 int     `json:"id" gorm:"primaryKey;autoIncrement"`
	RequestID          string  `json:"request_id" gorm:"not null"`                                           // Reference to data access request
	RequesterAccountID string  `json:"requester_account_id" gorm:"not null"`                                 // ID of the requester account
	RequesterAccount   Account `json:"requester_account" gorm:"foreignKey:RequesterAccountID;references:ID"` // Linked requester account                             // Automatically set creation timestamp
	ValidationStatus   string  `json:"validation_status" gorm:"not null;default:PENDING"`
	ValidationMessage  string  `json:"validation_message"`
}

func (m *DataAccessRequestRequester) TableName() string {
	return "data_access_request_requesters"
}

type DataAccessRequestRequesterTest struct {
	// Base requester information from data_access_request_requesters table
	ID                 int    `json:"id" gorm:"primaryKey;autoIncrement"`
	RequestID          string `json:"request_id" gorm:"not null"`
	RequesterAccountID string `json:"requester_account_id" gorm:"not null"`
	ValidationStatus   string `json:"validation_status" gorm:"not null;default:PENDING"`
	ValidationMessage  string `json:"validation_message"`

	// Relationships
	Request          DataAccessRequest `json:"request" gorm:"foreignKey:RequestID;references:ID"`
	RequesterAccount Account           `json:"requester_account" gorm:"foreignKey:RequesterAccountID;references:ID"`
}
