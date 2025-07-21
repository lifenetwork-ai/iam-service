package domain

import (
	"time"

	"github.com/google/uuid"
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

func (t *Tenant) ApplyTenantUpdate(name, publicURL, adminURL string) bool {
	updated := false

	if name != "" && name != t.Name {
		t.Name = name
		updated = true
	}
	if publicURL != "" && publicURL != t.PublicURL {
		t.PublicURL = publicURL
		updated = true
	}
	if adminURL != "" && adminURL != t.AdminURL {
		t.AdminURL = adminURL
		updated = true
	}
	if updated {
		t.UpdatedAt = time.Now().UTC()
	}
	return updated
}
