package ucases

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	repositories "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

type adminUseCase struct {
	tenantRepo       repositories.TenantRepository
	adminAccountRepo repositories.AdminAccountRepository
}

func NewAdminUseCase(tenantRepo repositories.TenantRepository, adminAccountRepo repositories.AdminAccountRepository) interfaces.AdminUseCase {
	return &adminUseCase{
		tenantRepo:       tenantRepo,
		adminAccountRepo: adminAccountRepo,
	}
}

func (u *adminUseCase) CreateAdminAccount(ctx context.Context, payload dto.CreateAdminAccountPayloadDTO) (*dto.AdminAccountDTO, *domainerrors.DomainError) {
	// Check if admin account with same email exists
	existingAccount, err := u.adminAccountRepo.GetByUsername(payload.Username)
	if err != nil {
		return nil, domainerrors.NewInternalError(
			"MSG_CREATE_ADMIN_FAILED",
			"Failed to check existing admin account",
		)
	}

	if existingAccount != nil {
		return nil, domainerrors.NewConflictError(
			"MSG_ADMIN_USERNAME_EXISTS",
			"Admin account with this username already exists",
			map[string]string{
				"field": "username",
				"error": "Username already exists",
			},
		)
	}

	// Create new admin account
	account := &domain.AdminAccount{}
	if err := account.FromCreateDTO(payload); err != nil {
		return nil, domainerrors.NewInternalError(
			"MSG_CREATE_ADMIN_FAILED",
			"Failed to create admin account",
		)
	}

	// Save to database
	if err := u.adminAccountRepo.Create(account); err != nil {
		logger.GetLogger().Errorf("Failed to save admin account: %v", err)
		return nil, domainerrors.NewInternalError(
			"MSG_CREATE_ADMIN_FAILED",
			"Failed to save admin account",
		)
	}

	// Return DTO
	dto := account.ToDTO()
	return &dto, nil
}

func (u *adminUseCase) GetAdminAccountByUsername(ctx context.Context, username string) (*dto.AdminAccountDTO, *domainerrors.DomainError) {
	account, err := u.adminAccountRepo.GetByUsername(username)
	if err != nil {
		return nil, domainerrors.NewInternalError(
			"MSG_GET_ADMIN_ACCOUNT_FAILED",
			"Failed to get admin account",
		)
	}

	if account == nil {
		return nil, domainerrors.NewNotFoundError(
			"MSG_ADMIN_ACCOUNT_NOT_FOUND",
			"Admin account not found",
		)
	}

	dto := account.ToDTO()
	return &dto, nil
}

func (u *adminUseCase) ListTenants(ctx context.Context, page, size int, keyword string) (*dto.PaginationDTOResponse, *domainerrors.DomainError) {
	tenants, err := u.tenantRepo.List()
	if err != nil {
		return nil, domainerrors.NewInternalError(
			"MSG_LIST_TENANTS_FAILED",
			"Failed to list tenants",
		)
	}

	// Filter by keyword if provided
	var filteredTenants []*domain.Tenant
	if keyword != "" {
		keyword = strings.ToLower(keyword)
		for _, tenant := range tenants {
			if strings.Contains(strings.ToLower(tenant.Name), keyword) {
				filteredTenants = append(filteredTenants, tenant)
			}
		}
	} else {
		filteredTenants = tenants
	}

	// Apply pagination
	start := (page - 1) * size
	end := start + size
	if start >= len(filteredTenants) {
		return &dto.PaginationDTOResponse{
			NextPage: page,
			Page:     page,
			Size:     size,
			Data:     []interface{}{},
		}, nil
	}
	if end > len(filteredTenants) {
		end = len(filteredTenants)
	}

	// Convert to DTOs
	tenantDTOs := make([]interface{}, 0)
	for _, tenant := range filteredTenants[start:end] {
		domainTenant := domain.Tenant{
			ID:        tenant.ID,
			Name:      tenant.Name,
			PublicURL: tenant.PublicURL,
			AdminURL:  tenant.AdminURL,
			CreatedAt: tenant.CreatedAt,
			UpdatedAt: tenant.UpdatedAt,
		}
		tenantDTOs = append(tenantDTOs, domainTenant.ToDTO())
	}

	nextPage := page
	if end < len(filteredTenants) {
		nextPage++
	}

	return &dto.PaginationDTOResponse{
		NextPage: nextPage,
		Page:     page,
		Size:     size,
		Data:     tenantDTOs,
	}, nil
}

func (u *adminUseCase) GetTenantByID(ctx context.Context, id string) (*dto.TenantDTO, *domainerrors.DomainError) {
	tenantID, err := uuid.Parse(id)
	if err != nil {
		return nil, domainerrors.NewValidationError(
			"MSG_INVALID_TENANT_ID_FORMAT",
			"Invalid tenant ID format",
			map[string]string{
				"field": "id",
				"error": "Invalid UUID format",
			},
		)
	}

	tenant, err := u.tenantRepo.GetByID(tenantID)
	if err != nil {
		return nil, domainerrors.NewInternalError(
			"MSG_GET_TENANT_FAILED",
			"Failed to get tenant",
		)
	}

	if tenant == nil {
		return nil, domainerrors.NewNotFoundError(
			"MSG_TENANT_NOT_FOUND",
			"Tenant not found",
		)
	}

	domainTenant := domain.Tenant{
		ID:        tenant.ID,
		Name:      tenant.Name,
		PublicURL: tenant.PublicURL,
		AdminURL:  tenant.AdminURL,
		CreatedAt: tenant.CreatedAt,
		UpdatedAt: tenant.UpdatedAt,
	}
	dto := domainTenant.ToDTO()
	return &dto, nil
}

func (u *adminUseCase) CreateTenant(ctx context.Context, payload dto.CreateTenantPayloadDTO) (*dto.TenantDTO, *domainerrors.DomainError) {
	// Check if tenant with same name exists
	existingTenant, err := u.tenantRepo.GetByName(payload.Name)
	if err != nil {
		return nil, domainerrors.NewInternalError(
			"MSG_CREATE_TENANT_FAILED",
			"Failed to create tenant",
		)
	}

	if existingTenant != nil {
		return nil, domainerrors.NewConflictError(
			"MSG_TENANT_NAME_EXISTS",
			fmt.Sprintf("Tenant with name '%s' already exists", payload.Name),
			map[string]string{
				"field": "name",
				"error": "Tenant name already exists",
			},
		)
	}

	// Create new tenant
	tenant := domain.FromCreateDTO(payload)
	repoTenant := &domain.Tenant{
		ID:        tenant.ID,
		Name:      tenant.Name,
		PublicURL: tenant.PublicURL,
		AdminURL:  tenant.AdminURL,
		CreatedAt: tenant.CreatedAt,
		UpdatedAt: tenant.UpdatedAt,
	}

	err = u.tenantRepo.Create(repoTenant)
	if err != nil {
		return nil, domainerrors.NewInternalError(
			"MSG_CREATE_TENANT_FAILED",
			"Failed to create tenant",
		)
	}

	dto := tenant.ToDTO()
	return &dto, nil
}

func (u *adminUseCase) UpdateTenant(ctx context.Context, id string, payload dto.UpdateTenantPayloadDTO) (*dto.TenantDTO, *domainerrors.DomainError) {
	tenantID, err := uuid.Parse(id)
	if err != nil {
		return nil, domainerrors.NewValidationError(
			"MSG_INVALID_TENANT_ID_FORMAT",
			"Invalid tenant ID format",
			map[string]string{
				"field": "id",
				"error": "Invalid UUID format",
			},
		)
	}

	// Get existing tenant
	existingTenant, err := u.tenantRepo.GetByID(tenantID)
	if err != nil {
		return nil, domainerrors.NewInternalError(
			"MSG_UPDATE_TENANT_FAILED",
			"Failed to update tenant",
		)
	}

	if existingTenant == nil {
		return nil, domainerrors.NewNotFoundError(
			"MSG_TENANT_NOT_FOUND",
			"Tenant not found",
		)
	}

	// Check name uniqueness if name is being updated
	if payload.Name != "" && payload.Name != existingTenant.Name {
		nameExists, err := u.tenantRepo.GetByName(payload.Name)
		if err != nil {
			return nil, domainerrors.NewInternalError(
				"MSG_UPDATE_TENANT_FAILED",
				"Failed to update tenant",
			)
		}

		if nameExists != nil {
			return nil, domainerrors.NewConflictError(
				"MSG_TENANT_NAME_EXISTS",
				fmt.Sprintf("Tenant with name '%s' already exists", payload.Name),
				map[string]string{
					"field": "name",
					"error": "Tenant name already exists",
				},
			)
		}
	}

	// Update tenant
	domainTenant := domain.Tenant{
		ID:        existingTenant.ID,
		Name:      existingTenant.Name,
		PublicURL: existingTenant.PublicURL,
		AdminURL:  existingTenant.AdminURL,
		CreatedAt: existingTenant.CreatedAt,
		UpdatedAt: existingTenant.UpdatedAt,
	}

	if domainTenant.ApplyUpdate(payload) {
		repoTenant := &domain.Tenant{
			ID:        domainTenant.ID,
			Name:      domainTenant.Name,
			PublicURL: domainTenant.PublicURL,
			AdminURL:  domainTenant.AdminURL,
			CreatedAt: domainTenant.CreatedAt,
			UpdatedAt: domainTenant.UpdatedAt,
		}

		err = u.tenantRepo.Update(repoTenant)
		if err != nil {
			return nil, domainerrors.NewInternalError(
				"MSG_UPDATE_TENANT_FAILED",
				"Failed to update tenant",
			)
		}
	}

	dto := domainTenant.ToDTO()
	return &dto, nil
}

func (u *adminUseCase) DeleteTenant(ctx context.Context, id string) (*dto.TenantDTO, *domainerrors.DomainError) {
	tenantID, err := uuid.Parse(id)
	if err != nil {
		return nil, domainerrors.NewValidationError(
			"MSG_INVALID_TENANT_ID_FORMAT",
			"Invalid tenant ID format",
			map[string]string{
				"field": "id",
				"error": "Invalid UUID format",
			},
		)
	}

	// Get existing tenant
	existingTenant, err := u.tenantRepo.GetByID(tenantID)
	if err != nil {
		return nil, domainerrors.NewInternalError(
			"MSG_DELETE_TENANT_FAILED",
			"Failed to delete tenant",
		)
	}

	if existingTenant == nil {
		return nil, domainerrors.NewNotFoundError(
			"MSG_TENANT_NOT_FOUND",
			"Tenant not found",
		)
	}

	// Convert to domain model for DTO conversion
	domainTenant := domain.Tenant{
		ID:        existingTenant.ID,
		Name:      existingTenant.Name,
		PublicURL: existingTenant.PublicURL,
		AdminURL:  existingTenant.AdminURL,
		CreatedAt: existingTenant.CreatedAt,
		UpdatedAt: existingTenant.UpdatedAt,
	}

	// Delete tenant
	err = u.tenantRepo.Delete(tenantID)
	if err != nil {
		return nil, domainerrors.NewInternalError(
			"MSG_DELETE_TENANT_FAILED",
			"Failed to delete tenant",
		)
	}

	dto := domainTenant.ToDTO()
	return &dto, nil
}

func (u *adminUseCase) UpdateTenantStatus(ctx context.Context, tenantID string, payload dto.UpdateTenantStatusPayloadDTO) (*dto.TenantDTO, *domainerrors.DomainError) {
	// TODO: Implement tenant status update
	return nil, domainerrors.NewInternalError(
		"MSG_UPDATE_TENANT_STATUS_NOT_IMPLEMENTED",
		"Tenant status update not implemented",
	)
}
