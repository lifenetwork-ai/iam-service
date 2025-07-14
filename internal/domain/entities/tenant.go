package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
)

// Tenant represents a tenant entity
type Tenant struct {
	ID        uuid.UUID
	Name      string
	PublicURL string
	AdminURL  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ToDTO converts a Tenant entity to a TenantDTO
func (t *Tenant) ToDTO() dto.TenantDTO {
	return dto.TenantDTO{
		ID:        t.ID.String(),
		Name:      t.Name,
		PublicURL: t.PublicURL,
		AdminURL:  t.AdminURL,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

// FromDTO creates a new Tenant entity from a CreateTenantPayloadDTO
func FromCreateDTO(payload dto.CreateTenantPayloadDTO) Tenant {
	return Tenant{
		ID:        uuid.New(),
		Name:      payload.Name,
		PublicURL: payload.PublicURL,
		AdminURL:  payload.AdminURL,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

// ApplyUpdate updates the Tenant entity from an UpdateTenantPayloadDTO
func (t *Tenant) ApplyUpdate(payload dto.UpdateTenantPayloadDTO) bool {
	updated := false

	if payload.Name != "" && payload.Name != t.Name {
		t.Name = payload.Name
		updated = true
	}

	if payload.PublicURL != "" && payload.PublicURL != t.PublicURL {
		t.PublicURL = payload.PublicURL
		updated = true
	}

	if payload.AdminURL != "" && payload.AdminURL != t.AdminURL {
		t.AdminURL = payload.AdminURL
		updated = true
	}

	if updated {
		t.UpdatedAt = time.Now().UTC()
	}

	return updated
}
