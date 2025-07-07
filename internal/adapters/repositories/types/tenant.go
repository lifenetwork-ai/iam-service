package interfaces

import (
	"time"

	"github.com/google/uuid"
)

// Tenant represents a tenant in the system
type Tenant struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	Name      string    `gorm:"type:varchar(255);not null"`
	PublicURL string    `gorm:"type:varchar(255);not null"`
	AdminURL  string    `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// TenantRepository defines the interface for tenant operations
type TenantRepository interface {
	Create(tenant *Tenant) error
	Update(tenant *Tenant) error
	Delete(id uuid.UUID) error
	GetByID(id uuid.UUID) (*Tenant, error)
	List() ([]*Tenant, error)
	GetByName(name string) (*Tenant, error)
}
