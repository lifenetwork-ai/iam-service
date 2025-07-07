package dto

import "time"

// AdminAccountDTO represents an admin account
type AdminAccountDTO struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateAdminAccountPayloadDTO represents the payload for creating an admin account
type CreateAdminAccountPayloadDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	// Role is always "admin" since only root can create admin accounts
	// and root account is configured via env vars
}

// UpdateTenantStatusPayloadDTO represents the payload for updating a tenant's status
type UpdateTenantStatusPayloadDTO struct {
	Status string `json:"status" binding:"required,oneof=active inactive suspended"`
	Reason string `json:"reason" binding:"required_if=Status suspended"`
}
