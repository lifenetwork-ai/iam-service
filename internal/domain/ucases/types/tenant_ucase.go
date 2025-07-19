package interfaces

import (
	"context"

	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
)

// TenantUseCase defines the interface for tenant use cases
type TenantUseCase interface {
	// List returns a paginated list of tenants
	List(ctx context.Context, page, size int, keyword string) (*dto.PaginationDTOResponse, *domainerrors.DomainError)

	// GetByID returns a tenant by ID
	GetByID(ctx context.Context, id string) (*dto.TenantDTO, *domainerrors.DomainError)

	// Create creates a new tenant
	Create(ctx context.Context, payload dto.CreateTenantPayloadDTO) (*dto.TenantDTO, *domainerrors.DomainError)

	// Update updates an existing tenant
	Update(ctx context.Context, id string, payload dto.UpdateTenantPayloadDTO) (*dto.TenantDTO, *domainerrors.DomainError)

	// Delete deletes a tenant
	Delete(ctx context.Context, id string) (*dto.TenantDTO, *domainerrors.DomainError)
}
