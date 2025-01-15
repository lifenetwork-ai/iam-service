package domain

type DataAccessRequestRequester struct {
	ID                 int     `json:"id" gorm:"primaryKey;autoIncrement"`
	RequestID          string  `json:"request_id" gorm:"not null"`                                           // Reference to data access request
	RequesterAccountID string  `json:"requester_account_id" gorm:"not null"`                                 // ID of the requester account
	RequesterAccount   Account `json:"requester_account" gorm:"foreignKey:RequesterAccountID;references:ID"` // Linked requester account                             // Automatically set creation timestamp
}

func (m *DataAccessRequestRequester) TableName() string {
	return "data_access_request_requesters"
}
