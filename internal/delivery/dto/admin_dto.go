package dto

import "time"

// AdminAccountDTO represents an admin account
type AdminAccountDTO struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateAdminAccountPayloadDTO represents the payload for creating an admin account
type CreateAdminAccountPayloadDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required,oneof=admin"`
}

// UpdateTenantStatusPayloadDTO represents the payload for updating a tenant's status
type UpdateTenantStatusPayloadDTO struct {
	Status string `json:"status" binding:"required,oneof=active inactive suspended"`
	Reason string `json:"reason" binding:"required_if=Status suspended"`
}
