package dto

import (
	"time"

	"github.com/google/uuid"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

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

// FromCreateDTO creates domain entity AdminAccount from CreateAdminAccountPayloadDTO
func FromCreateDTO(payload CreateAdminAccountPayloadDTO) domain.AdminAccount {
	a := domain.AdminAccount{
		ID:           uuid.New(),
		Username:     payload.Username,
		Role:         payload.Role,
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		PasswordHash: payload.Password,
	}

	return a
}

// ToDTO converts domain entity AdminAccount to AdminAccountDTO
func ToAdminAccountDTO(a domain.AdminAccount) AdminAccountDTO {
	return AdminAccountDTO{
		ID:        a.ID.String(),
		Username:  a.Username,
		Role:      a.Role,
		Status:    a.Status,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

type AdminAddIdentifierPayloadDTO struct {
	ExistingIdentifier string `json:"existing_identifier" binding:"required"`
	NewIdentifier      string `json:"new_identifier" binding:"required"`
}
