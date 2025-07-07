package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	httpresponse "github.com/lifenetwork-ai/iam-service/packages/http/response"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

type adminHandler struct {
	ucase interfaces.AdminUseCase
}

func NewAdminHandler(ucase interfaces.AdminUseCase) *adminHandler {
	return &adminHandler{
		ucase: ucase,
	}
}

// CreateAdminAccount creates a new admin account.
// @Summary Create a new admin account
// @Security BasicAuth
// @Description Create a new admin account (requires root account configured via ADMIN_EMAIL and ADMIN_PASSWORD env vars)
// @Tags admin
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param Authorization header string true "Bearer Token"
// @Param admin body dto.CreateAdminAccountPayloadDTO true "Admin account details"
// @Success 201 {object} dto.AdminAccountDTO "Successful creation of admin account"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 401 {object} response.ErrorResponse "Unauthorized - Not the root account"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/admin/accounts [post]
func (h *adminHandler) CreateAdminAccount(ctx *gin.Context) {
	var reqPayload dto.CreateAdminAccountPayloadDTO

	if err := ctx.ShouldBindJSON(&reqPayload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_PAYLOAD",
			"Invalid request payload",
			err,
		)
		return
	}

	response, errResponse := h.ucase.CreateAdminAccount(ctx, reqPayload)
	if errResponse != nil {
		httpresponse.Error(
			ctx,
			errResponse.Status,
			errResponse.Code,
			errResponse.Message,
			errResponse.Details,
		)
		return
	}

	httpresponse.Success(ctx, http.StatusCreated, response)
}

// ListTenants returns a paginated list of tenants
// @Summary List all tenants
// @Security BasicAuth
// @Description Get a paginated list of tenants with optional search
// @Tags tenants
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param size query int false "Page size (default: 10)"
// @Param keyword query string false "Search keyword"
// @Success 200 {object} dto.PaginationDTOResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/admin/tenants [get]
func (h *adminHandler) ListTenants(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(ctx.DefaultQuery("size", "10"))
	keyword := ctx.Query("keyword")

	response, errResponse := h.ucase.ListTenants(ctx, page, size, keyword)
	if errResponse != nil {
		httpresponse.Error(
			ctx,
			errResponse.Status,
			errResponse.Code,
			errResponse.Message,
			errResponse.Details,
		)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, response)
}

// GetTenant returns a tenant by ID
// @Summary Get a tenant by ID
// @Security BasicAuth
// @Description Get detailed information about a tenant
// @Tags tenants
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {object} dto.TenantDTO
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/admin/tenants/{id} [get]
func (h *adminHandler) GetTenant(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_TENANT_ID",
			"Invalid tenant ID",
			nil,
		)
		return
	}

	response, errResponse := h.ucase.GetTenantByID(ctx, id)
	if errResponse != nil {
		httpresponse.Error(
			ctx,
			errResponse.Status,
			errResponse.Code,
			errResponse.Message,
			errResponse.Details,
		)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, response)
}

// CreateTenant creates a new tenant
// @Summary Create a new tenant
// @Description Create a new tenant with the provided details
// @Tags tenants
// @Accept json
// @Produce json
// @Param tenant body dto.CreateTenantPayloadDTO true "Tenant details"
// @Success 201 {object} dto.TenantDTO
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Router /api/v1/admin/tenants [post]
func (h *adminHandler) CreateTenant(ctx *gin.Context) {
	var payload dto.CreateTenantPayloadDTO
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_PAYLOAD",
			"Invalid request payload",
			err,
		)
		return
	}

	response, errResponse := h.ucase.CreateTenant(ctx, payload)
	if errResponse != nil {
		httpresponse.Error(
			ctx,
			errResponse.Status,
			errResponse.Code,
			errResponse.Message,
			errResponse.Details,
		)
		return
	}

	httpresponse.Success(ctx, http.StatusCreated, response)
}

// UpdateTenant updates an existing tenant
// @Summary Update a tenant
// @Description Update a tenant's details
// @Tags tenants
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Param tenant body dto.UpdateTenantPayloadDTO true "Tenant details"
// @Success 200 {object} dto.TenantDTO
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Router /api/v1/admin/tenants/{id} [put]
func (h *adminHandler) UpdateTenant(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_TENANT_ID",
			"Invalid tenant ID",
			nil,
		)
		return
	}

	var payload dto.UpdateTenantPayloadDTO
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_PAYLOAD",
			"Invalid request payload",
			err,
		)
		return
	}

	response, errResponse := h.ucase.UpdateTenant(ctx, id, payload)
	if errResponse != nil {
		httpresponse.Error(
			ctx,
			errResponse.Status,
			errResponse.Code,
			errResponse.Message,
			errResponse.Details,
		)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, response)
}

// DeleteTenant deletes a tenant
// @Summary Delete a tenant
// @Description Delete a tenant and all associated data
// @Tags tenants
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {object} dto.TenantDTO
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/admin/tenants/{id} [delete]
func (h *adminHandler) DeleteTenant(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_TENANT_ID",
			"Invalid tenant ID",
			nil,
		)
		return
	}

	response, errResponse := h.ucase.DeleteTenant(ctx, id)
	if errResponse != nil {
		httpresponse.Error(
			ctx,
			errResponse.Status,
			errResponse.Code,
			errResponse.Message,
			errResponse.Details,
		)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, response)
}

// UpdateTenantStatus updates a tenant's status
// @Summary Update tenant status
// @Description Update a tenant's status (requires root or admin account)
// @Tags tenants
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Param status body dto.UpdateTenantStatusPayloadDTO true "Tenant status details"
// @Success 200 {object} dto.TenantDTO
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/admin/tenants/{id}/status [put]
func (h *adminHandler) UpdateTenantStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_TENANT_ID",
			"Invalid tenant ID",
			nil,
		)
		return
	}

	var payload dto.UpdateTenantStatusPayloadDTO
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_PAYLOAD",
			"Invalid request payload",
			err,
		)
		return
	}

	response, errResponse := h.ucase.UpdateTenantStatus(ctx, id, payload)
	if errResponse != nil {
		httpresponse.Error(
			ctx,
			errResponse.Status,
			errResponse.Code,
			errResponse.Message,
			errResponse.Details,
		)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, response)
}
