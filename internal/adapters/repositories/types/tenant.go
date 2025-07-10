package interfaces

import (
	"github.com/google/uuid"
	entities "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

// TenantRepository defines the interface for tenant operations
type TenantRepository interface {
	Create(tenant *entities.Tenant) error
	Update(tenant *entities.Tenant) error
	Delete(id uuid.UUID) error
	GetByID(id uuid.UUID) (*entities.Tenant, error)
	List() ([]*entities.Tenant, error)
	GetByName(name string) (*entities.Tenant, error)
}
