package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	dto "github.com/genefriendway/human-network-iam/internal/delivery/dto"
	interfaces "github.com/genefriendway/human-network-iam/internal/domain/ucases/types"
	httpresponse "github.com/genefriendway/human-network-iam/packages/http/response"
	"github.com/genefriendway/human-network-iam/packages/logger"
)

type roleHandler struct {
	ucase interfaces.IdentityRoleUseCase
}

func NewIdentityRoleHandler(ucase interfaces.IdentityRoleUseCase) *roleHandler {
	return &roleHandler{
		ucase: ucase,
	}
}

// GetRoles retrieves a list of roles.
// @Summary Retrieve roles
// @Description Get roles
// @Tags roles
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param Authorization header string true "Bearer Token"
// @Param page query int false "Page number"
// @Param size query int false "Page size"
// @Param keyword query string false "Keyword"
// @Success 200 {object} dto.PaginationDTOResponse "Successful retrieval of roles"
// @Failure 400 {object} response.ErrorResponse "Invalid page number or size"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/roles [get]
func (h *roleHandler) GetRoles(ctx *gin.Context) {
	page := ctx.DefaultQuery("page", "1")
	size := ctx.DefaultQuery("size", "10")
	keyword := ctx.DefaultQuery("keyword", "")

	// Parse page and size into integers
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		logger.GetLogger().Errorf("Invalid page number: %v", err)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_PAGE_NUMBER",
			"Invalid page number",
			err,
		)
		return
	}

	sizeInt, err := strconv.Atoi(size)
	if err != nil || sizeInt < 1 {
		logger.GetLogger().Errorf("Invalid size: %v", err)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_SIZE",
			"Invalid size",
			err,
		)
		return
	}

	response, errResponse := h.ucase.List(ctx, pageInt, sizeInt, keyword)
	if errResponse != nil {
		logger.GetLogger().Errorf("Failed to get roles: %v", errResponse)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_FAILED_GET_ROLES",
			"Failed to get roles",
			errResponse,
		)
		return
	}

	// Return the response as a JSON response
	ctx.JSON(http.StatusOK, response)
}

// GetDetail retrieves a role by it's ID.
// @Summary Retrieve role by ID
// @Description Get role by ID
// @Tags roles
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param Authorization header string true "Bearer Token"
// @Param role_id path string true "role ID"
// @Success 200 {object} dto.IdentityRoleDTO "Successful retrieval of role"
// @Failure 400 {object} response.ErrorResponse "Invalid request ID"
// @Failure 404 {object} response.ErrorResponse "role not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/roles/{role_id} [get]
func (h *roleHandler) GetDetail(ctx *gin.Context) {
	// Extract and parse role_id from query string
	roleId := ctx.Query("role_id")
	if roleId == "" {
		logger.GetLogger().Error("Invalid role ID")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	role, err := h.ucase.GetByID(ctx, roleId)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get role: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role"})
		return
	}

	// Return the response as a JSON response
	ctx.JSON(http.StatusOK, role)
}

// CreateRole creates a new role.
// @Summary Create a new role
// @Description Create a new role
// @Tags roles
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param Authorization header string true "Bearer Token"
// @Param role body dto.CreateIdentityRolePayloadDTO true "role payload"
// @Success 201 {object} dto.IdentityRoleDTO "Successful creation of role"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/roles [post]
func (h *roleHandler) CreateRole(ctx *gin.Context) {
	var reqPayload dto.CreateIdentityRolePayloadDTO

	// Parse and validate the request payload
	if err := ctx.ShouldBindJSON(&reqPayload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_PAGE_NUMBER",
			"Invalid page number",
			err,
		)
		return
	}

	// Create the role
	response, err := h.ucase.Create(ctx, reqPayload)
	if err != nil {
		logger.GetLogger().Errorf("Failed to create role: %v", err)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_SIZE",
			"Invalid size",
			err,
		)
		return
	}

	// Return the response as a JSON response
	ctx.JSON(http.StatusCreated, response)
}

// UpdateRole updates an existing role.
// @Summary Update an existing role
// @Description Update an existing role
// @Tags roles
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param Authorization header string true "Bearer Token"
// @Param role_id path string true "role ID"
// @Param role body dto.UpdateIdentityRolePayloadDTO true "role payload"
// @Success 200 {object} dto.IdentityRoleDTO "Successful update of role"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 404 {object} response.ErrorResponse "role not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/roles/{role_id} [put]
func (h *roleHandler) UpdateRole(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}

// DeleteRole deletes an existing role.
// @Summary Delete an existing role
// @Description Delete an existing role
// @Tags roles
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param Authorization header string true "Bearer Token"
// @Param role_id path string true "role ID"
// @Success 204 "Successful deletion of role"
// @Failure 400 {object} response.ErrorResponse "Invalid request ID"
// @Failure 404 {object} response.ErrorResponse "role not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/roles/{role_id} [delete]
func (h *roleHandler) DeleteRole(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}
