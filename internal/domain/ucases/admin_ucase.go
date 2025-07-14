package ucases

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	repositories "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
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

func (u *adminUseCase) CreateAdminAccount(ctx context.Context, payload dto.CreateAdminAccountPayloadDTO) (*dto.AdminAccountDTO, *dto.ErrorDTOResponse) {
	// Check if admin account with same email exists
	existingAccount, err := u.adminAccountRepo.GetByUsername(payload.Username)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_CREATE_ADMIN_FAILED",
			Message: "Failed to check existing admin account",
			Details: []interface{}{err.Error()},
		}
	}

	if existingAccount != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusConflict,
			Code:    "MSG_ADMIN_USERNAME_EXISTS",
			Message: "Admin account with this username already exists",
			Details: []interface{}{
				map[string]string{
					"field": "username",
					"error": "Username already exists",
				},
			},
		}
	}

	// Create new admin account
	account := &domain.AdminAccount{}
	if err := account.FromCreateDTO(payload); err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_CREATE_ADMIN_FAILED",
			Message: "Failed to create admin account",
			Details: []interface{}{err.Error()},
		}
	}

	// Save to database
	if err := u.adminAccountRepo.Create(account); err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_CREATE_ADMIN_FAILED",
			Message: "Failed to save admin account",
			Details: []interface{}{err.Error()},
		}
	}

	// Return DTO
	dto := account.ToDTO()
	return &dto, nil
}

func (u *adminUseCase) GetAdminAccountByUsername(ctx context.Context, username string) (*dto.AdminAccountDTO, *dto.ErrorDTOResponse) {
	account, err := u.adminAccountRepo.GetByUsername(username)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_GET_ADMIN_ACCOUNT_FAILED",
			Message: err.Error(),
			Details: []interface{}{err.Error()},
		}
	}

	if account == nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusNotFound,
			Code:    "MSG_ADMIN_ACCOUNT_NOT_FOUND",
			Message: "Admin account not found",
			Details: []interface{}{"Admin account not found"},
		}
	}

	dto := account.ToDTO()
	return &dto, nil
}

func (u *adminUseCase) ListTenants(ctx context.Context, page, size int, keyword string) (*dto.PaginationDTOResponse, *dto.ErrorDTOResponse) {
	tenants, err := u.tenantRepo.List()
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
			Details: []interface{}{err.Error()},
		}
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

func (u *adminUseCase) GetTenantByID(ctx context.Context, id string) (*dto.TenantDTO, *dto.ErrorDTOResponse) {
	tenantID, err := uuid.Parse(id)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid tenant ID format",
			Details: []interface{}{"Invalid UUID format"},
		}
	}

	tenant, err := u.tenantRepo.GetByID(tenantID)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
			Details: []interface{}{err.Error()},
		}
	}

	if tenant == nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusNotFound,
			Code:    "MSG_TENANT_NOT_FOUND",
			Message: "Tenant not found",
			Details: []interface{}{
				map[string]string{"field": "id", "error": "Tenant not found"},
			},
		}
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

func (u *adminUseCase) CreateTenant(ctx context.Context, payload dto.CreateTenantPayloadDTO) (*dto.TenantDTO, *dto.ErrorDTOResponse) {
	// Check if tenant with same name exists
	existingTenant, err := u.tenantRepo.GetByName(payload.Name)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
			Details: []interface{}{err.Error()},
		}
	}

	if existingTenant != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusConflict,
			Message: fmt.Sprintf("Tenant with name '%s' already exists", payload.Name),
			Details: []interface{}{
				map[string]string{"field": "name", "error": "Tenant name already exists"},
			},
		}
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
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
			Details: []interface{}{err.Error()},
		}
	}

	dto := tenant.ToDTO()
	return &dto, nil
}

func (u *adminUseCase) UpdateTenant(ctx context.Context, id string, payload dto.UpdateTenantPayloadDTO) (*dto.TenantDTO, *dto.ErrorDTOResponse) {
	tenantID, err := uuid.Parse(id)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid tenant ID format",
			Details: []interface{}{"Invalid UUID format"},
		}
	}

	// Get existing tenant
	existingTenant, err := u.tenantRepo.GetByID(tenantID)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
			Details: []interface{}{err.Error()},
		}
	}

	if existingTenant == nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusNotFound,
			Code:    "MSG_TENANT_NOT_FOUND",
			Message: "Tenant not found",
			Details: []interface{}{
				map[string]string{"field": "id", "error": "Tenant not found"},
			},
		}
	}

	// Check name uniqueness if name is being updated
	if payload.Name != "" && payload.Name != existingTenant.Name {
		nameExists, err := u.tenantRepo.GetByName(payload.Name)
		if err != nil {
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusInternalServerError,
				Message: err.Error(),
				Details: []interface{}{err.Error()},
			}
		}

		if nameExists != nil {
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusConflict,
				Message: fmt.Sprintf("Tenant with name '%s' already exists", payload.Name),
				Details: []interface{}{
					map[string]string{"field": "name", "error": "Tenant name already exists"},
				},
			}
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
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusInternalServerError,
				Message: err.Error(),
				Details: []interface{}{err.Error()},
			}
		}
	}

	dto := domainTenant.ToDTO()
	return &dto, nil
}

func (u *adminUseCase) DeleteTenant(ctx context.Context, id string) (*dto.TenantDTO, *dto.ErrorDTOResponse) {
	tenantID, err := uuid.Parse(id)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid tenant ID format",
			Details: []interface{}{"Invalid UUID format"},
		}
	}

	// Get existing tenant
	existingTenant, err := u.tenantRepo.GetByID(tenantID)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
			Details: []interface{}{err.Error()},
		}
	}

	if existingTenant == nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusNotFound,
			Code:    "MSG_TENANT_NOT_FOUND",
			Message: "Tenant not found",
			Details: []interface{}{
				map[string]string{"field": "id", "error": "Tenant not found"},
			},
		}
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
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
			Details: []interface{}{err.Error()},
		}
	}

	dto := domainTenant.ToDTO()
	return &dto, nil
}

func (u *adminUseCase) UpdateTenantStatus(ctx context.Context, tenantID string, payload dto.UpdateTenantStatusPayloadDTO) (*dto.TenantDTO, *dto.ErrorDTOResponse) {
	// TODO: Implement tenant status update
	return nil, &dto.ErrorDTOResponse{
		Status:  http.StatusNotImplemented,
		Message: "Tenant status update not implemented",
	}
}
