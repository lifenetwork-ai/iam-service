package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	"github.com/lifenetwork-ai/iam-service/internal/delivery/http/middleware"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	httpresponse "github.com/lifenetwork-ai/iam-service/packages/http/response"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

type permissionHandler struct {
	ucase interfaces.PermissionUseCase
}

func NewPermissionHandler(ucase interfaces.PermissionUseCase) *permissionHandler {
	return &permissionHandler{
		ucase: ucase,
	}
}

// CreateRelationTuple creates a relation tuple
// @Summary Create relation tuple
// @Security BasicAuth
// @Description Create a relation tuple for a tenant member
// @Tags permissions
// @Accept json
// @Produce json
// @Param X-Tenant-Id header string true "Tenant ID"
// @Param Authorization header string true "Bearer Token (Bearer ory...)" default(Bearer <token>)
// @Param request body dto.CreateRelationTupleRequestDTO true "Relation tuple creation request"
// @Success 200 {object} response.SuccessResponse "Relation tuple created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/permissions/relation-tuples [post]
func (h *permissionHandler) CreateRelationTuple(c *gin.Context) {
	_, err := middleware.GetTenantFromContext(c)
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

	tenant, err := middleware.GetTenantFromContext(c)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get tenant from context: %v", err)
		httpresponse.Error(
			c,
			http.StatusBadRequest,
			"MSG_INVALID_TENANT",
			"Invalid tenant",
			err,
		)
		return
	}

	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get user profile: %v", err)
		httpresponse.Error(
			c,
			http.StatusInternalServerError,
			"MSG_GET_USER_PROFILE_FAILED",
			"Failed to get user profile",
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

	// Validate request payload
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

	usecaseReq := interfaces.CreateRelationTupleRequest{
		Namespace: req.Namespace,
		Relation:  req.Relation,
		Object:    req.Object,
		SubjectSet: interfaces.TenantRelation{
			TenantID: tenant.ID.String(),
			UserID:   user.ID,
		},
	}

	ucaseErr := h.ucase.CreateRelationTuple(c.Request.Context(), usecaseReq)
	if ucaseErr != nil {
		logger.GetLogger().Errorf("Failed to create relation tuple: %v", err)
		httpresponse.Error(
			c,
			http.StatusInternalServerError,
			"MSG_CREATE_RELATION_TUPLE_FAILED",
			"Failed to create relation tuple",
			ucaseErr,
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
// @Success 200 {object} dto.CheckPermissionResponseDTO "Permission check result"
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/permissions/check [post]
func (h *permissionHandler) CheckPermission(c *gin.Context) {
	tenant, err := middleware.GetTenantFromContext(c)
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

	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get user profile: %v", err)
		httpresponse.Error(
			c,
			http.StatusInternalServerError,
			"MSG_GET_USER_PROFILE_FAILED",
			"Failed to get user profile",
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

	ucaseReq := interfaces.CheckPermissionRequest{
		Namespace: req.Namespace,
		Relation:  req.Relation,
		Object:    req.Object,
		TenantRelation: interfaces.TenantRelation{
			TenantID: tenant.ID.String(),
			UserID:   user.ID,
		},
	}

	// Check permission using Keto
	allowed, ucaseErr := h.ucase.CheckPermission(c.Request.Context(), ucaseReq)
	if ucaseErr != nil {
		logger.GetLogger().Errorf("Failed to check permission: %v", ucaseErr)
		httpresponse.Error(
			c,
			http.StatusInternalServerError,
			"MSG_PERMISSION_CHECK_FAILED",
			"Failed to check permission",
			ucaseErr,
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
