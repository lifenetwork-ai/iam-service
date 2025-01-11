package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

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
// @Tags policy
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
