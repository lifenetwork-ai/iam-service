package ucases

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domaintypes "github.com/lifenetwork-ai/iam-service/internal/domain/types"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
)

type tenantUseCase struct {
	tenantRepo domainrepo.TenantRepository
}

func NewTenantUseCase(tenantRepo domainrepo.TenantRepository) interfaces.TenantUseCase {
	return &tenantUseCase{
		tenantRepo: tenantRepo,
	}
}

// GetAll retrieves all tenants
func (u *tenantUseCase) GetAll(ctx context.Context) ([]*domain.Tenant, *domainerrors.DomainError) {
	tenants, err := u.tenantRepo.List()
	if err != nil {
		return nil, domainerrors.NewInternalError(
			"MSG_GET_TENANT_LIST_FAILED",
			"Failed to get tenant list",
		)
	}

	return tenants, nil
}

func (u *tenantUseCase) List(
	ctx context.Context,
	page int,
	size int,
	keyword string,
) (*domaintypes.PaginatedResponse[domain.Tenant], *domainerrors.DomainError) {
	tenants, err := u.tenantRepo.List()
	if err != nil {
		return nil, domainerrors.NewInternalError(
			"MSG_GET_TENANT_LIST_FAILED",
			"Failed to get tenant list",
		)
	}

	// Filter by keyword if provided
	var filteredTenants []domain.Tenant
	if keyword != "" {
		keyword = strings.ToLower(keyword)
		for _, tenant := range tenants {
			if strings.Contains(strings.ToLower(tenant.Name), keyword) {
				filteredTenants = append(filteredTenants, *tenant)
			}
		}
	} else {
		filteredTenants = make([]domain.Tenant, 0, len(tenants))
		for _, tenant := range tenants {
			filteredTenants = append(filteredTenants, *tenant)
		}
	}

	return domaintypes.CalculatePagination(filteredTenants, page, size), nil
}

func (u *tenantUseCase) GetByID(
	ctx context.Context,
	id string,
) (*domain.Tenant, *domainerrors.DomainError) {
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

	domainTenant := domain.Tenant{
		ID:        tenant.ID,
		Name:      tenant.Name,
		PublicURL: tenant.PublicURL,
		AdminURL:  tenant.AdminURL,
		CreatedAt: tenant.CreatedAt,
		UpdatedAt: tenant.UpdatedAt,
	}
	return &domainTenant, nil
}

func (u *tenantUseCase) Create(
	ctx context.Context,
	name, publicURL, adminURL string,
) (*domain.Tenant, *domainerrors.DomainError) {
	// Check if tenant with same name exists
	existingTenant, err := u.tenantRepo.GetByName(name)
	if err != nil {
		return nil, domainerrors.NewInternalError(
			"MSG_CREATE_TENANT_FAILED",
			"Failed to create tenant",
		)
	}

	if existingTenant != nil {
		return nil, domainerrors.NewConflictError(
			"MSG_TENANT_ALREADY_EXISTS",
			fmt.Sprintf("Tenant with name '%s' already exists", name),
			map[string]string{"field": "name", "error": "Tenant name already exists"},
		)
	}

	// Create new tenant
	tenant := domain.Tenant{
		Name:      name,
		PublicURL: publicURL,
		AdminURL:  adminURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
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

	return &tenant, nil
}

func (u *tenantUseCase) Update(
	ctx context.Context,
	id string,
	name, publicURL, adminURL string,
) (*domain.Tenant, *domainerrors.DomainError) {
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
	if name != "" && name != existingTenant.Name {
		nameExists, err := u.tenantRepo.GetByName(name)
		if err != nil {
			return nil, domainerrors.NewInternalError(
				"MSG_GET_TENANT_BY_NAME_FAILED",
				"Failed to get tenant by name",
			)
		}

		if nameExists != nil {
			return nil, domainerrors.NewConflictError(
				"MSG_TENANT_NAME_ALREADY_EXISTS",
				fmt.Sprintf("Tenant with name '%s' already exists", name),
				map[string]string{"field": "name", "error": "Tenant name already exists"},
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

	if domainTenant.ApplyTenantUpdate(name, publicURL, adminURL) {
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

	return &domainTenant, nil
}

func (u *tenantUseCase) Delete(
	ctx context.Context,
	id string,
) (*domain.Tenant, *domainerrors.DomainError) {
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

	return &domainTenant, nil
}
