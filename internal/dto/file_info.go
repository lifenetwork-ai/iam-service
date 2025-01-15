package dto

import "time"

type FileInfoDTO struct {
	ID         string    `json:"id"`          // Unique identifier for the file
	Name       string    `json:"name"`        // File name
	ShareCount int       `json:"share_count"` // Number of shares
	OwnerID    string    `json:"owner_id"`    // Owner ID
	CreatedAt  time.Time `json:"created_at"`  // Creation timestamp
	UpdatedAt  time.Time `json:"updated_at"`  // Last updated timestamp
}
