package ucases

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domaintypes "github.com/lifenetwork-ai/iam-service/internal/domain/types"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

type adminUseCase struct {
	tenantRepo       domainrepo.TenantRepository
	adminAccountRepo domainrepo.AdminAccountRepository
}

func NewAdminUseCase(tenantRepo domainrepo.TenantRepository, adminAccountRepo domainrepo.AdminAccountRepository) interfaces.AdminUseCase {
	return &adminUseCase{
		tenantRepo:       tenantRepo,
		adminAccountRepo: adminAccountRepo,
	}
}

func (u *adminUseCase) CreateAdminAccount(ctx context.Context, username, password, role string) (*domain.AdminAccount, *domainerrors.DomainError) {
	// Check if admin account with same email exists
	existingAccount, err := u.adminAccountRepo.GetByUsername(username)
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

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.GetLogger().Errorf("Failed to hash password: %v", err)
		return nil, domainerrors.NewInternalError(
			"MSG_CREATE_ADMIN_FAILED",
			"Failed to hash admin password",
		)
	}

	// Create new admin account
	account := domain.AdminAccount{
		Username:     username,
		Name:         "",
		PasswordHash: string(hashedPassword),
		Role:         role,
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save to database
	if err := u.adminAccountRepo.Create(&account); err != nil {
		logger.GetLogger().Errorf("Failed to save admin account: %v", err)
		return nil, domainerrors.NewInternalError(
			"MSG_CREATE_ADMIN_FAILED",
			"Failed to save admin account",
		)
	}

	// Return DTO
	return &account, nil
}

func (u *adminUseCase) GetAdminAccountByUsername(ctx context.Context, username string) (*domain.AdminAccount, *domainerrors.DomainError) {
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

	return account, nil
}

func (u *adminUseCase) ListTenants(ctx context.Context, page, size int, keyword string) (*domaintypes.PaginatedResponse[*domain.Tenant], *domainerrors.DomainError) {
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
	return domaintypes.CalculatePagination(filteredTenants, page, size), nil
}

func (u *adminUseCase) GetTenantByID(ctx context.Context, id string) (*domain.Tenant, *domainerrors.DomainError) {
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
	return &domainTenant, nil
}

func (u *adminUseCase) CreateTenant(ctx context.Context, name, publicURL, adminURL string) (*domain.Tenant, *domainerrors.DomainError) {
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
			"MSG_TENANT_NAME_EXISTS",
			fmt.Sprintf("Tenant with name '%s' already exists", name),
			map[string]string{
				"field": "name",
				"error": "Tenant name already exists",
			},
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

func (u *adminUseCase) UpdateTenant(ctx context.Context, id, name, publicURL, adminURL string) (*domain.Tenant, *domainerrors.DomainError) {
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
	if name != "" && name != existingTenant.Name {
		nameExists, err := u.tenantRepo.GetByName(name)
		if err != nil {
			return nil, domainerrors.NewInternalError(
				"MSG_UPDATE_TENANT_FAILED",
				"Failed to update tenant",
			)
		}

		if nameExists != nil {
			return nil, domainerrors.NewConflictError(
				"MSG_TENANT_NAME_EXISTS",
				fmt.Sprintf("Tenant with name '%s' already exists", name),
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

func (u *adminUseCase) DeleteTenant(ctx context.Context, id string) (*domain.Tenant, *domainerrors.DomainError) {
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

	return &domainTenant, nil
}
