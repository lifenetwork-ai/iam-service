package interfaces

import (
	"context"

	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	"github.com/lifenetwork-ai/iam-service/internal/domain/types"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
)

// AdminUseCase defines the interface for administrative operations
type AdminUseCase interface {
	// CreateAdminAccount creates a new admin account (requires root account configured via env vars)
	CreateAdminAccount(ctx context.Context, username, password, role string) (*domain.AdminAccount, *domainerrors.DomainError)
	GetAdminAccountByUsername(ctx context.Context, username string) (*domain.AdminAccount, *domainerrors.DomainError)
	// Tenant Management
	ListTenants(ctx context.Context, page, size int, keyword string) (*types.PaginatedResponse[*domain.Tenant], *domainerrors.DomainError)
	GetTenantByID(ctx context.Context, id string) (*domain.Tenant, *domainerrors.DomainError)
	CreateTenant(ctx context.Context, name, publicURL, adminURL string) (*domain.Tenant, *domainerrors.DomainError)
	UpdateTenant(ctx context.Context, id, name, publicURL, adminURL string) (*domain.Tenant, *domainerrors.DomainError)
	DeleteTenant(ctx context.Context, id string) (*domain.Tenant, *domainerrors.DomainError)
}
