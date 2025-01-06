package domain

import "time"

type DataAccessRequest struct {
	ID                 string    `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID             string    `json:"user_id" gorm:"not null"`                          // User receiving the request
	CustomerID         string    `json:"customer_id" gorm:"not null"`                      // Customer requesting access
	ReasonForRequest   string    `json:"reason_for_request" gorm:"not null"`               // Reason for the request
	Status             string    `json:"status" gorm:"type:varchar(20);default:'PENDING'"` // Request status
	ReasonForRejection *string   `json:"reason_for_rejection,omitempty"`                   // Reason for user rejection (optional)
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
