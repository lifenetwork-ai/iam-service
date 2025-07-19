package ucases

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	repositories "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	entities "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
)

type tenantUseCase struct {
	tenantRepo repositories.TenantRepository
}

func NewTenantUseCase(tenantRepo repositories.TenantRepository) interfaces.TenantUseCase {
	return &tenantUseCase{
		tenantRepo: tenantRepo,
	}
}

func (u *tenantUseCase) List(
	ctx context.Context,
	page int,
	size int,
	keyword string,
) (*dto.PaginationDTOResponse, *domainerrors.DomainError) {
	tenants, err := u.tenantRepo.List()
	if err != nil {
		return nil, domainerrors.NewInternalError(
			"MSG_GET_TENANT_LIST_FAILED",
			"Failed to get tenant list",
		)
	}

	// Filter by keyword if provided
	var filteredTenants []*entities.Tenant
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
		domainTenant := entities.Tenant{
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

func (u *tenantUseCase) GetByID(
	ctx context.Context,
	id string,
) (*dto.TenantDTO, *domainerrors.DomainError) {
	tenantID, err := uuid.Parse(id)
	if err != nil {
		return nil, domainerrors.NewValidationError(
			"MSG_INVALID_TENANT_ID_FORMAT",
			"Invalid tenant ID format",
			map[string]string{"field": "id", "error": "Invalid UUID format"},
		)
	}

	tenant, err := u.tenantRepo.GetByID(tenantID)
	if err != nil {
		return nil, domainerrors.NewInternalError(
			"MSG_GET_TENANT_BY_ID_FAILED",
			"Failed to get tenant by ID",
		)
	}

	if tenant == nil {
		return nil, domainerrors.NewNotFoundError(
			"MSG_TENANT_NOT_FOUND",
			"Tenant not found",
		)
	}

	domainTenant := entities.Tenant{
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

func (u *tenantUseCase) Create(
	ctx context.Context,
	payload dto.CreateTenantPayloadDTO,
) (*dto.TenantDTO, *domainerrors.DomainError) {
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
			"MSG_TENANT_ALREADY_EXISTS",
			fmt.Sprintf("Tenant with name '%s' already exists", payload.Name),
			map[string]string{"field": "name", "error": "Tenant name already exists"},
		)
	}

	// Create new tenant
	tenant := entities.FromCreateDTO(payload)
	repoTenant := &entities.Tenant{
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

func (u *tenantUseCase) Update(
	ctx context.Context,
	id string,
	payload dto.UpdateTenantPayloadDTO,
) (*dto.TenantDTO, *domainerrors.DomainError) {
	tenantID, err := uuid.Parse(id)
	if err != nil {
		return nil, domainerrors.NewValidationError(
			"MSG_INVALID_TENANT_ID_FORMAT",
			"Invalid tenant ID format",
			map[string]string{"field": "id", "error": "Invalid UUID format"},
		)
	}

	// Get existing tenant
	existingTenant, err := u.tenantRepo.GetByID(tenantID)
	if err != nil {
		return nil, domainerrors.NewInternalError(
			"MSG_GET_TENANT_BY_ID_FAILED",
			"Failed to get tenant by ID",
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
				"MSG_GET_TENANT_BY_NAME_FAILED",
				"Failed to get tenant by name",
			)
		}

		if nameExists != nil {
			return nil, domainerrors.NewConflictError(
				"MSG_TENANT_NAME_ALREADY_EXISTS",
				fmt.Sprintf("Tenant with name '%s' already exists", payload.Name),
				map[string]string{"field": "name", "error": "Tenant name already exists"},
			)
		}
	}

	// Update tenant
	domainTenant := entities.Tenant{
		ID:        existingTenant.ID,
		Name:      existingTenant.Name,
		PublicURL: existingTenant.PublicURL,
		AdminURL:  existingTenant.AdminURL,
		CreatedAt: existingTenant.CreatedAt,
		UpdatedAt: existingTenant.UpdatedAt,
	}

	if domainTenant.ApplyUpdate(payload) {
		repoTenant := &entities.Tenant{
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

func (u *tenantUseCase) Delete(
	ctx context.Context,
	id string,
) (*dto.TenantDTO, *domainerrors.DomainError) {
	tenantID, err := uuid.Parse(id)
	if err != nil {
		return nil, domainerrors.NewValidationError(
			"MSG_INVALID_TENANT_ID_FORMAT",
			"Invalid tenant ID format",
			map[string]string{"field": "id", "error": "Invalid UUID format"},
		)
	}

	// Get existing tenant
	existingTenant, err := u.tenantRepo.GetByID(tenantID)
	if err != nil {
		return nil, domainerrors.NewInternalError(
			"MSG_GET_TENANT_BY_ID_FAILED",
			"Failed to get tenant by ID",
		)
	}

	if existingTenant == nil {
		return nil, domainerrors.NewNotFoundError(
			"MSG_TENANT_NOT_FOUND",
			"Tenant not found",
		)
	}

	// Convert to domain model for DTO conversion
	domainTenant := entities.Tenant{
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
