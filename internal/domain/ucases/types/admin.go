package interfaces

import (
	"context"

	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
)

// AdminUseCase defines the interface for administrative operations
type AdminUseCase interface {
	// CreateAdminAccount creates a new admin account (requires root account configured via env vars)
	CreateAdminAccount(ctx context.Context, payload dto.CreateAdminAccountPayloadDTO) (*dto.AdminAccountDTO, *dto.ErrorDTOResponse)
	GetAdminAccountByUsername(ctx context.Context, username string) (*dto.AdminAccountDTO, *dto.ErrorDTOResponse)
	// Tenant Management
	ListTenants(ctx context.Context, page, size int, keyword string) (*dto.PaginationDTOResponse, *dto.ErrorDTOResponse)
	GetTenantByID(ctx context.Context, id string) (*dto.TenantDTO, *dto.ErrorDTOResponse)
	CreateTenant(ctx context.Context, payload dto.CreateTenantPayloadDTO) (*dto.TenantDTO, *dto.ErrorDTOResponse)
	UpdateTenant(ctx context.Context, id string, payload dto.UpdateTenantPayloadDTO) (*dto.TenantDTO, *dto.ErrorDTOResponse)
	DeleteTenant(ctx context.Context, id string) (*dto.TenantDTO, *dto.ErrorDTOResponse)
	UpdateTenantStatus(ctx context.Context, tenantID string, payload dto.UpdateTenantStatusPayloadDTO) (*dto.TenantDTO, *dto.ErrorDTOResponse)
}
