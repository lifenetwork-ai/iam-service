package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
	httpresponse "github.com/genefriendway/human-network-auth/pkg/http/response"
	"github.com/genefriendway/human-network-auth/pkg/logger"
)

type accountHandler struct {
	accountUCase interfaces.AccountUCase
	authUCase    interfaces.AuthUCase
}

func NewAccountHandler(
	accountUCase interfaces.AccountUCase,
	authUCase interfaces.AuthUCase,
) *accountHandler {
	return &accountHandler{
		accountUCase: accountUCase,
		authUCase:    authUCase,
	}
}

// GetCurrentUser retrieves the currently authenticated user's details.
// @Summary Get current user details
// @Description This endpoint retrieves the details of the currently authenticated user using the provided access token.
// @Tags account
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Success 200 {object} dto.AccountDetailDTO "User details"
// @Failure 401 {object} response.GeneralError "Invalid or expired token"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/account/me [get]
func (h *accountHandler) GetCurrentUser(ctx *gin.Context) {
	// Retrieve the token from the context
	token, exists := ctx.Get("token")
	if !exists {
		httpresponse.Error(ctx, http.StatusUnauthorized, "Token not found", nil)
		return
	}

	// Validate the token and fetch user details
	account, err := h.authUCase.ValidateToken(token.(string))
	if err != nil {
		logger.GetLogger().Errorf("Failed to validate token: %v", err)
		httpresponse.Error(ctx, http.StatusUnauthorized, "Invalid or expired token", err)
		return
	}

	// Fetch role-specific details
	detail, err := h.accountUCase.FindDetailByAccountID(account.ID, constants.AccountRole(account.Role))
	if err != nil {
		logger.GetLogger().Errorf("Failed to fetch user details: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to fetch user details", err)
		return
	}

	// Respond with user details
	ctx.JSON(http.StatusOK, detail)
}

// GetActiveValidators retrieves the list of active validators.
// @Summary Get Active Validators
// @Description Fetches a list of active validators.
// @Tags validators
// @Accept json
// @Produce json
// @Success 200 {array} dto.AccountDetailDTO "List of active validators"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/validators/active [get]
func (h *accountHandler) GetActiveValidators(ctx *gin.Context) {
	// Fetch active validators using the use case
	validators, err := h.accountUCase.GetActiveValidators()
	if err != nil {
		logger.GetLogger().Errorf("Failed to fetch active validators: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to fetch active validators", err)
		return
	}

	// Respond with the list of active validators
	ctx.JSON(http.StatusOK, gin.H{"validators": validators})
}
