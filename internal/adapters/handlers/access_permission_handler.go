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

type permissionHandler struct {
	ucase interfaces.AccessPermissionUseCase
}

func NewAccessPermissionHandler(ucase interfaces.AccessPermissionUseCase) *permissionHandler {
	return &permissionHandler{
		ucase: ucase,
	}
}

// GetPermissions retrieves a list of permissions.
// @Summary Retrieve permissions
// @Description Get permissions
// @Tags permissions
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param size query int false "Page size"
// @Param keyword query string false "Keyword"
// @Success 200 {object} dto.PaginationDTOResponse "Successful retrieval of permissions"
// @Failure 400 {object} response.ErrorResponse "Invalid page number or size"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/permissions [get]
func (h *permissionHandler) GetPermissions(ctx *gin.Context) {
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
		logger.GetLogger().Errorf("Failed to get permissions: %v", errResponse)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_FAILED_TO_GET_PERMISSIONS",
			"Failed to get permissions",
			errResponse,
		)
		return
	}

	// Return the response as a JSON response
	httpresponse.Success(ctx, http.StatusOK, response)
}

// GetDetail retrieves a permission by it's ID.
// @Summary Retrieve permission by ID
// @Description Get permission by ID
// @Tags permissions
// @Accept json
// @Produce json
// @Param permission_id path string true "permission ID"
// @Success 200 {object} dto.AccessPermissionDTO "Successful retrieval of permission"
// @Failure 400 {object} response.ErrorResponse "Invalid request ID"
// @Failure 404 {object} response.ErrorResponse "permission not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/permissions/{permission_id} [get]
func (h *permissionHandler) GetDetail(ctx *gin.Context) {
	// Extract and parse permission_id from query string
	permissionId := ctx.Query("permission_id")
	if permissionId == "" {
		logger.GetLogger().Error("Invalid permission ID")
		httpresponse.Error(ctx, http.StatusBadRequest, "MSG_INVALID_PERMISSION_ID", "Invalid permission ID", nil)
		return
	}

	permission, err := h.ucase.GetByID(ctx, permissionId)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get permission: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "MSG_FAILED_TO_GET_PERMISSION", "Failed to get permission", err)
		return
	}

	// Return the response as a JSON response
	httpresponse.Success(ctx, http.StatusOK, permission)
}

// CreatePermission creates a new permission.
// @Summary Create a new permission
// @Description Create a new permission
// @Tags permissions
// @Accept json
// @Produce json
// @Param permission body dto.CreateAccessPermissionPayloadDTO true "permission payload"
// @Success 201 {object} dto.AccessPermissionDTO "Successful creation of permission"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/permissions [post]
func (h *permissionHandler) CreatePermission(ctx *gin.Context) {
	var reqPayload dto.CreateAccessPermissionPayloadDTO

	// Parse and validate the request payload
	if err := ctx.ShouldBindJSON(&reqPayload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "MSG_INVALID_PAYLOAD", "Invalid request payload", err)
		return
	}

	// Create the permission
	response, err := h.ucase.Create(ctx, reqPayload)
	if err != nil {
		logger.GetLogger().Errorf("Failed to create permission: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "MSG_FAILED_TO_CREATE_PERMISSION", "Failed to create permission", err)
		return
	}

	// Return the response as a JSON response
	httpresponse.Success(ctx, http.StatusCreated, response)
}

// UpdatePermission updates an existing permission.
// @Summary Update an existing permission
// @Description Update an existing permission
// @Tags permissions
// @Accept json
// @Produce json
// @Param permission_id path string true "permission ID"
// @Param permission body dto.UpdateAccessPermissionPayloadDTO true "permission payload"
// @Success 200 {object} dto.AccessPermissionDTO "Successful update of permission"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 404 {object} response.ErrorResponse "permission not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/permissions/{permission_id} [put]
func (h *permissionHandler) UpdatePermission(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}

// DeletePermission deletes an existing permission.
// @Summary Delete an existing permission
// @Description Delete an existing permission
// @Tags permissions
// @Accept json
// @Produce json
// @Param permission_id path string true "permission ID"
// @Success 204 "Successful deletion of permission"
// @Failure 400 {object} response.ErrorResponse "Invalid request ID"
// @Failure 404 {object} response.ErrorResponse "permission not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/permissions/{permission_id} [delete]
func (h *permissionHandler) DeletePermission(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}
