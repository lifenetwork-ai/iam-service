package dto

import "time"

// PermissionDTO represents the Data Transfer Object for a Permission.
type PermissionDTO struct {
	ID          string    `json:"id"`          // Unique ID for the permission
	PolicyID    string    `json:"policy_id"`   // Foreign key referencing IAMPolicy
	Resource    string    `json:"resource"`    // The resource this permission applies to
	Action      string    `json:"action"`      // The action this permission allows
	Description string    `json:"description"` // Optional description of the permission
	CreatedAt   time.Time `json:"created_at"`  // Timestamp of permission creation
	UpdatedAt   time.Time `json:"updated_at"`  // Timestamp of last update
}
