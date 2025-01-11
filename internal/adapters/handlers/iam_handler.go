package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
	"github.com/genefriendway/human-network-auth/pkg/http/response"
	"github.com/genefriendway/human-network-auth/pkg/logger"
)

type iamHandler struct {
	iamUCase interfaces.IAMUCase
}

func NewIAMHandler(iamUCase interfaces.IAMUCase) *iamHandler {
	return &iamHandler{iamUCase: iamUCase}
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
	// Parse and validate the payload
	var payload dto.PolicyPayloadDTO
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		response.Error(ctx, http.StatusBadRequest, "Invalid payload", err)
		return
	}

	// Call the use case to create the policy
	policy, err := h.iamUCase.CreatePolicy(payload)
	if err != nil {
		logger.GetLogger().Errorf("Failed to create policy: %v", err)
		response.Error(ctx, http.StatusInternalServerError, "Failed to create policy", err)
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
	accountID := ctx.Param("accountID")
	if accountID == "" {
		response.Error(ctx, http.StatusBadRequest, "Account ID is required", nil)
		return
	}

	var payload dto.AssignPolicyPayloadDTO
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		response.Error(ctx, http.StatusBadRequest, "Invalid payload", err)
		return
	}

	err := h.iamUCase.AssignPolicyToAccount(accountID, payload.PolicyID)
	if err != nil {
		if errors.Is(err, domain.ErrDataNotFound) {
			response.Error(ctx, http.StatusNotFound, "Account or policy not found", nil)
		} else {
			logger.GetLogger().Errorf("Failed to assign policy: %v", err)
			response.Error(ctx, http.StatusInternalServerError, "Failed to assign policy", err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Policy assigned successfully"})
}
