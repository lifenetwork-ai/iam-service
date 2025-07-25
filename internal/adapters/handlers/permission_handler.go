package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	"github.com/lifenetwork-ai/iam-service/internal/delivery/http/middleware"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
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

	_, err = middleware.GetUserFromContext(c)
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

	// // Check if the user has permission to create relation tuples
	// checkReq := types.CheckPermissionRequest{
	// 	Namespace: req.Namespace,
	// 	Relation:  "manage",
	// 	Object:    req.Object,
	// 	TenantRelation: types.TenantRelation{
	// 		TenantID:   tenant.ID.String(),
	// 		Identifier: user.Email,
	// 	},
	// }

	// canManage, ucaseErr := h.ucase.CheckPermission(c.Request.Context(), checkReq)
	// if ucaseErr != nil {
	// 	logger.GetLogger().Errorf("Failed to check management permission: %v", ucaseErr)
	// 	httpresponse.Error(
	// 		c,
	// 		http.StatusInternalServerError,
	// 		"MSG_CHECK_MANAGEMENT_PERMISSION_FAILED",
	// 		"Failed to check management permission",
	// 		ucaseErr,
	// 	)
	// 	return
	// }

	// if !canManage {
	// 	httpresponse.Error(
	// 		c,
	// 		http.StatusForbidden,
	// 		"MSG_MANAGEMENT_NOT_ALLOWED",
	// 		"You don't have permission to manage relation tuples for this resource",
	// 		nil,
	// 	)
	// 	return
	// }

	usecaseReq := types.CreateRelationTupleRequest{
		Namespace: req.Namespace,
		Relation:  req.Relation,
		Object:    req.Object,
		TenantRelation: types.TenantRelation{
			TenantID:   tenant.ID.String(),
			Identifier: req.Identifier,
		},
	}

	ucaseErr := h.ucase.CreateRelationTuple(c.Request.Context(), usecaseReq)
	if ucaseErr != nil {
		logger.GetLogger().Errorf("Failed to create relation tuple: %v", ucaseErr)
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

// SelfCheckPermission checks if a subject has permission to perform an action on an object
// @Summary User-facing permission check
// @Description Check if a subject has permission to perform an action on an object
// @Tags permissions
// @Accept json
// @Produce json
// @Param X-Tenant-Id header string true "Tenant ID"
// @Param request body dto.SelfCheckPermissionRequestDTO true "Permission check request"
// @Success 200 {object} dto.CheckPermissionResponseDTO "Permission check result"
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/permissions/self-check [post]
func (h *permissionHandler) SelfCheckPermission(c *gin.Context) {
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

	var req dto.SelfCheckPermissionRequestDTO
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

	identifier := user.Email
	if identifier == "" {
		identifier = user.Phone
	}

	ucaseReq := types.CheckPermissionRequest{
		Namespace: req.Namespace,
		Relation:  req.Relation,
		Object:    req.Object,
		TenantRelation: types.TenantRelation{
			TenantID:   tenant.ID.String(),
			Identifier: identifier,
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

// DelegateAccess delegates access to a resource
// @Summary Delegate access
// @Description Delegate access to a resource
// @Tags permissions
// @Accept json
// @Produce json
// @Param X-Tenant-Id header string true "Tenant ID"
// @Param Authorization header string true "Bearer Token (Bearer ory...)" default(Bearer <token>)
// @Param request body dto.DelegateAccessRequestDTO true "Delegate access request"
// @Success 200 {object} response.SuccessResponse "Access delegated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/permissions/delegate [post]
func (h *permissionHandler) DelegateAccess(c *gin.Context) {
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

	var req dto.DelegateAccessRequestDTO
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

	// Check if the delegator has permission to delegate
	checkReq := types.CheckPermissionRequest{
		Namespace: req.ResourceType,
		Relation:  "delegate",
		Object:    fmt.Sprintf("%s:%s", req.ResourceType, req.ResourceID),
		TenantRelation: types.TenantRelation{
			TenantID:   tenant.ID.String(),
			Identifier: user.Email,
		},
	}

	canDelegate, ucaseErr := h.ucase.CheckPermission(c.Request.Context(), checkReq)
	if ucaseErr != nil {
		logger.GetLogger().Errorf("Failed to check delegation permission: %v", ucaseErr)
		httpresponse.Error(
			c,
			http.StatusInternalServerError,
			"MSG_CHECK_DELEGATION_PERMISSION_FAILED",
			"Failed to check delegation permission",
			ucaseErr,
		)
		return
	}

	if !canDelegate {
		httpresponse.Error(
			c,
			http.StatusForbidden,
			"MSG_DELEGATION_NOT_ALLOWED",
			"You don't have permission to delegate access to this resource",
			nil,
		)
		return
	}

	ucaseReq := types.DelegateAccessRequest{
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Permission:   req.Permission,
		TenantID:     tenant.ID.String(),
		Identifier:   req.Identifier,
	}

	allowed, ucaseErr := h.ucase.DelegateAccess(c.Request.Context(), ucaseReq)
	if ucaseErr != nil {
		logger.GetLogger().Errorf("Failed to delegate access: %v", ucaseErr)
		httpresponse.Error(
			c,
			http.StatusInternalServerError,
			"MSG_DELEGATE_ACCESS_FAILED",
			"Failed to delegate access",
			ucaseErr,
		)
		return
	}

	httpresponse.Success(c, http.StatusOK, allowed)
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

	ucaseReq := types.CheckPermissionRequest{
		Namespace: req.Namespace,
		Relation:  req.Relation,
		Object:    req.Object,
		TenantRelation: types.TenantRelation{
			TenantID:   tenant.ID.String(),
			Identifier: req.TenantMember.Identifier,
		},
	}

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

	response := dto.CheckPermissionResponseDTO{
		Allowed: allowed,
	}
	if !allowed {
		response.Reason = "Permission denied by policy"
	}

	httpresponse.Success(c, http.StatusOK, response)
}
