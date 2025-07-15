package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/keto"
	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	"github.com/lifenetwork-ai/iam-service/internal/delivery/http/middleware"
	httpresponse "github.com/lifenetwork-ai/iam-service/packages/http/response"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

type permissionHandler struct {
	ketoClient *keto.Client
}

func NewPermissionHandler(ketoClient *keto.Client) *permissionHandler {
	return &permissionHandler{
		ketoClient: ketoClient,
	}
}

func (h *permissionHandler) CreateRelationTuple(c *gin.Context) {
	_, err := middleware.GetTenantFromContext(c.Request.Context())
	if err != nil {
		logger.GetLogger().Errorf("Failed to get tenant: %v", err)
		httpresponse.Error(
			c,
			http.StatusBadRequest,
			"MSG_INVALID_TENANT",
			"Invalid tenant",
			err,
		)
		return
	}

	var req dto.CreateRelationTupleRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(
			c,
			http.StatusBadRequest,
			"MSG_INVALID_PAYLOAD",
			"Invalid request payload",
			err,
		)
		return
	}

	// Sanitize request
	if err := req.Validate(); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(
			c,
			http.StatusBadRequest,
			"MSG_INVALID_PAYLOAD",
			"Invalid request payload",
			err,
		)
		return
	}

	err = h.ketoClient.CreateRelationTuple(c.Request.Context(), req)
	if err != nil {
		logger.GetLogger().Errorf("Failed to create relation tuple: %v", err)
		httpresponse.Error(
			c,
			http.StatusInternalServerError,
			"MSG_CREATE_RELATION_TUPLE_FAILED",
			"Failed to create relation tuple",
			err,
		)
		return
	}

	httpresponse.Success(c, http.StatusOK, "Relation tuple created successfully")
}

// CheckPermission checks if a subject has permission to perform an action on an object
// @Summary Check permission
// @Description Check if a subject has permission to perform an action on an object
// @Tags permissions
// @Accept json
// @Produce json
// @Param X-Tenant-Id header string true "Tenant ID"
// @Param request body dto.CheckPermissionRequestDTO true "Permission check request"
// @Success 200 {object} dto.CheckPermissionResponseDTO
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/permissions/check [post]
func (h *permissionHandler) CheckPermission(c *gin.Context) {
	_, err := middleware.GetTenantFromContext(c.Request.Context())
	if err != nil {
		logger.GetLogger().Errorf("Failed to get tenant: %v", err)
		httpresponse.Error(
			c,
			http.StatusBadRequest,
			"MSG_INVALID_TENANT",
			"Invalid tenant",
			err,
		)
		return
	}

	var req dto.CheckPermissionRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(
			c,
			http.StatusBadRequest,
			"MSG_INVALID_PAYLOAD",
			"Invalid request payload",
			err,
		)
		return
	}

	if err := req.Validate(); err != nil {

		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(
			c,
			http.StatusBadRequest,
			"MSG_INVALID_PAYLOAD",
			"Invalid request payload",
			err,
		)
		return
	}

	// Check permission using Keto
	allowed, err := h.ketoClient.CheckPermission(c.Request.Context(), req)
	if err != nil {
		logger.GetLogger().Errorf("Failed to check permission: %v", err)
		httpresponse.Error(
			c,
			http.StatusInternalServerError,
			"MSG_PERMISSION_CHECK_FAILED",
			"Failed to check permission",
			err,
		)
		return
	}

	// Return response
	response := dto.CheckPermissionResponseDTO{
		Allowed: allowed,
	}
	if !allowed {
		response.Reason = "Permission denied by policy"
	}

	httpresponse.Success(c, http.StatusOK, response)
}
