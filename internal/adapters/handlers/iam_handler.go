package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/dto"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
	httpresponse "github.com/genefriendway/human-network-iam/pkg/http/response"
	"github.com/genefriendway/human-network-iam/pkg/logger"
)

type iamHandler struct {
	iamUCase  interfaces.IAMUCase
	authUCase interfaces.AuthUCase
}

func NewIAMHandler(iamUCase interfaces.IAMUCase, authUCase interfaces.AuthUCase) *iamHandler {
	return &iamHandler{
		iamUCase:  iamUCase,
		authUCase: authUCase,
	}
}

// CreatePolicy creates a new policy.
// @Summary Create a new policy
// @Description Adds a new IAM policy to the system. Only accessible to Admins.
// @Tags IAM
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param payload body dto.PolicyPayloadDTO true "Payload for creating policy"
// @Success 201 {object} map[string]interface{} "Policy created successfully"
// @Failure 400 {object} response.GeneralError "Invalid payload"
// @Failure 401 {object} response.GeneralError "Unauthorized"
// @Failure 403 {object} response.GeneralError "Forbidden"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/iam/policies [post]
func (h *iamHandler) CreatePolicy(ctx *gin.Context) {
	// Retrieve the authenticated account from the context
	_, exists := ctx.Get("account")
	if !exists {
		httpresponse.Error(ctx, http.StatusUnauthorized, "Unauthorized access: account information missing", nil)
		return
	}

	// Parse and validate the payload
	var payload dto.PolicyPayloadDTO
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid payload", err)
		return
	}

	// Call the use case to create the policy
	policy, err := h.iamUCase.CreatePolicy(payload)
	if err != nil {
		logger.GetLogger().Errorf("Failed to create policy: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to create policy", err)
		return
	}

	// Respond with success
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Policy created successfully",
		"policy":  policy,
	})
}

// AssignPolicyToAccount assigns a policy to an account.
// @Summary Assign a policy to an account
// @Description Maps a specified policy to an account by accountID.
// @Tags IAM
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token (e.g., 'Bearer <token>')"
// @Param accountID path string true "Account ID to assign the policy to"
// @Param payload body dto.AssignPolicyPayloadDTO true "Payload containing the policy ID"
// @Success 200 {object} map[string]interface{} "Policy assigned successfully"
// @Failure 400 {object} response.GeneralError "Invalid payload or missing policy"
// @Failure 404 {object} response.GeneralError "Account or policy not found"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/iam/accounts/{accountID}/policies [post]
func (h *iamHandler) AssignPolicyToAccount(ctx *gin.Context) {
	// Retrieve the authenticated account from the context
	_, exists := ctx.Get("account")
	if !exists {
		httpresponse.Error(ctx, http.StatusUnauthorized, "Unauthorized access: account information missing", nil)
		return
	}

	// Parse the account ID from the request
	accountID := ctx.Param("accountID")
	if accountID == "" {
		httpresponse.Error(ctx, http.StatusBadRequest, "Account ID is required", nil)
		return
	}

	var payload dto.AssignPolicyPayloadDTO
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid payload", err)
		return
	}

	err := h.iamUCase.AssignPolicyToAccount(accountID, payload.PolicyID)
	if err != nil {
		if errors.Is(err, domain.ErrDataNotFound) {
			httpresponse.Error(ctx, http.StatusNotFound, "Account or policy not found", nil)
		} else {
			logger.GetLogger().Errorf("Failed to assign policy: %v", err)
			httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to assign policy", err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Policy assigned successfully"})
}

// GetPoliciesWithPermissions retrieves all policies and their associated permissions.
// @Summary Get policies with permissions
// @Description Fetches a list of IAM policies along with their associated permissions.
// @Tags IAM
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Success 200 {array} dto.PolicyWithPermissionsDTO "List of policies with permissions"
// @Failure 401 {object} response.GeneralError "Unauthorized"
// @Failure 403 {object} response.GeneralError "Insufficient permissions"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/iam/policies [get]
func (h *iamHandler) GetPoliciesWithPermissions(ctx *gin.Context) {
	// Retrieve the authenticated account from the context
	_, exists := ctx.Get("account")
	if !exists {
		httpresponse.Error(ctx, http.StatusUnauthorized, "Unauthorized access: account information missing", nil)
		return
	}

	// Fetch policies with permissions
	policiesWithPermissions, err := h.iamUCase.GetPoliciesWithPermissions()
	if err != nil {
		logger.GetLogger().Errorf("Failed to fetch policies with permissions: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to fetch policies and permissions", err)
		return
	}

	// Respond with the data
	ctx.JSON(http.StatusOK, policiesWithPermissions)
}

// AssignPermissionToPolicy assigns a permission to an existing policy.
// @Summary Assign a permission to a policy
// @Description Adds a new permission to an existing IAM policy.
// @Tags IAM
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param payload body dto.PermissionPayloadDTO true "Payload for assigning permission"
// @Success 201 {object} map[string]interface{} "Permission assigned successfully"
// @Failure 400 {object} response.GeneralError "Invalid payload or missing policy"
// @Failure 404 {object} response.GeneralError "Policy not found"
// @Failure 409 {object} response.GeneralError "Permission already exists"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/iam/policies/permissions [post]
func (h *iamHandler) AssignPermissionToPolicy(ctx *gin.Context) {
	// Retrieve the authenticated account from the context
	_, exists := ctx.Get("account")
	if !exists {
		httpresponse.Error(ctx, http.StatusUnauthorized, "Unauthorized access: account information missing", nil)
		return
	}

	// Parse and validate the payload
	var payload dto.PermissionPayloadDTO
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid payload format", err)
		return
	}

	// Validate resource and action
	if payload.Resource == "" || payload.Action == "" {
		httpresponse.Error(ctx, http.StatusBadRequest, "Resource and action are required", nil)
		return
	}

	// Add the permission to the policy
	err := h.iamUCase.CreatePermission(dto.PermissionPayloadDTO{
		PolicyID:    payload.PolicyID,
		Resource:    payload.Resource,
		Action:      payload.Action,
		Description: payload.Description,
	})
	if err != nil {
		if errors.Is(err, domain.ErrDataNotFound) {
			httpresponse.Error(ctx, http.StatusNotFound, "Policy not found", nil)
		} else if errors.Is(err, domain.ErrAlreadyExists) {
			httpresponse.Error(ctx, http.StatusConflict, "Permission already exists for this policy", nil)
		} else {
			logger.GetLogger().Errorf("Failed to assign permission: %v", err)
			httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to assign permission", err)
		}
		return
	}

	// Respond with success
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Permission assigned successfully",
	})
}
