package interfaces

import (
	"context"

	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
)

// AdminUseCase defines the interface for administrative operations
type AdminUseCase interface {
	// CreateAdminAccount creates a new admin account (requires root account configured via env vars)
	CreateAdminAccount(ctx context.Context, payload dto.CreateAdminAccountPayloadDTO) (*dto.AdminAccountDTO, *domainerrors.DomainError)
	GetAdminAccountByUsername(ctx context.Context, username string) (*dto.AdminAccountDTO, *domainerrors.DomainError)
	// Tenant Management
	ListTenants(ctx context.Context, page, size int, keyword string) (*dto.PaginationDTOResponse, *domainerrors.DomainError)
	GetTenantByID(ctx context.Context, id string) (*dto.TenantDTO, *domainerrors.DomainError)
	CreateTenant(ctx context.Context, payload dto.CreateTenantPayloadDTO) (*dto.TenantDTO, *domainerrors.DomainError)
	UpdateTenant(ctx context.Context, id string, payload dto.UpdateTenantPayloadDTO) (*dto.TenantDTO, *domainerrors.DomainError)
	DeleteTenant(ctx context.Context, id string) (*dto.TenantDTO, *domainerrors.DomainError)
	UpdateTenantStatus(ctx context.Context, tenantID string, payload dto.UpdateTenantStatusPayloadDTO) (*dto.TenantDTO, *domainerrors.DomainError)
}
