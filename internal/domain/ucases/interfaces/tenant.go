package interfaces

import (
	"context"

	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domaintypes "github.com/lifenetwork-ai/iam-service/internal/domain/types"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
)

// TenantUseCase defines the interface for tenant use cases
type TenantUseCase interface {
	// GetAll retrieves all tenants
	GetAll(ctx context.Context) ([]*domain.Tenant, *domainerrors.DomainError)

	// List returns a paginated list of tenants
	List(ctx context.Context, page, size int, keyword string) (*domaintypes.PaginatedResponse[domain.Tenant], *domainerrors.DomainError)

	// GetByID returns a tenant by ID
	GetByID(ctx context.Context, id string) (*domain.Tenant, *domainerrors.DomainError)

	// Create creates a new tenant
	Create(ctx context.Context, name, publicURL, adminURL string) (*domain.Tenant, *domainerrors.DomainError)

	// Update updates an existing tenant
	Update(ctx context.Context, id, name, publicURL, adminURL string) (*domain.Tenant, *domainerrors.DomainError)

	// Delete deletes a tenant
	Delete(ctx context.Context, id string) (*domain.Tenant, *domainerrors.DomainError)
}
