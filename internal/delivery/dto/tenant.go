package dto

import (
	"time"

	"github.com/google/uuid"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

// TenantDTO represents the tenant data transfer object
type TenantDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	PublicURL string    `json:"public_url"`
	AdminURL  string    `json:"admin_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateTenantPayloadDTO represents the payload for creating a tenant
type CreateTenantPayloadDTO struct {
	Name      string `json:"name" validate:"required"`
	PublicURL string `json:"public_url" validate:"required,url"`
	AdminURL  string `json:"admin_url" validate:"required,url"`
}

// UpdateTenantPayloadDTO represents the payload for updating a tenant
type UpdateTenantPayloadDTO struct {
	Name      string `json:"name" validate:"omitempty"`
	PublicURL string `json:"public_url" validate:"omitempty,url"`
	AdminURL  string `json:"admin_url" validate:"omitempty,url"`
}

func ToTenantDTO(t domain.Tenant) TenantDTO {
	return TenantDTO{
		ID:        t.ID.String(),
		Name:      t.Name,
		PublicURL: t.PublicURL,
		AdminURL:  t.AdminURL,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func FromCreateTenantPayloadDTO(payload CreateTenantPayloadDTO) domain.Tenant {
	return domain.Tenant{
		ID:        uuid.New(),
		Name:      payload.Name,
		PublicURL: payload.PublicURL,
		AdminURL:  payload.AdminURL,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}
