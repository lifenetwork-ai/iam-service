package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
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
// @Param Authorization header string true "Bearer access token (e.g., 'Bearer <token>')"
// @Success 200 {object} dto.AccountDetailDTO "User details"
// @Failure 401 {object} response.GeneralError "Unauthorized"
// @Failure 403 {object} response.GeneralError "Forbidden"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/account/me [get]
func (h *accountHandler) GetCurrentUser(ctx *gin.Context) {
	// Retrieve the token from the context
	token, exists := ctx.Get("token")
	if !exists {
		httpresponse.Error(ctx, http.StatusUnauthorized, "Token not found", domain.ErrTokenNotFound)
		return
	}

	// Validate the token and fetch user details
	account, err := h.authUCase.ValidateToken(token.(string))
	if err != nil {
		logger.GetLogger().Errorf("Failed to validate token: %v", err)
		httpresponse.Error(ctx, http.StatusUnauthorized, "Invalid token", err)
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
// @Param Authorization header string true "Bearer access token (e.g., 'Bearer <token>')"
// @Success 200 {array} dto.AccountDetailDTO "List of active validators"
// @Failure 401 {object} response.GeneralError "Unauthorized"
// @Failure 403 {object} response.GeneralError "Forbidden"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/validators/active [get]
func (h *accountHandler) GetActiveValidators(ctx *gin.Context) {
	// Retrieve the token from the context
	token, exists := ctx.Get("token")
	if !exists {
		httpresponse.Error(ctx, http.StatusUnauthorized, "Token not found", domain.ErrTokenNotFound)
		return
	}

	// Validate the token and fetch user details
	account, err := h.authUCase.ValidateToken(token.(string))
	if err != nil {
		logger.GetLogger().Errorf("Failed to validate token: %v", err)
		httpresponse.Error(ctx, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	// Ensure the user has the required permissions
	if account.Role != string(constants.User) {
		httpresponse.Error(ctx, http.StatusForbidden, "Insufficient permissions", domain.ErrInsufficientPermissions)
		return
	}

	// Fetch active validators
	validators, err := h.accountUCase.GetActiveValidators()
	if err != nil {
		logger.GetLogger().Errorf("Failed to fetch active validators: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to fetch active validators", err)
		return
	}

	// Return the list of active validators
	ctx.JSON(http.StatusOK, gin.H{"validators": validators})
}

// UpdateAccountRole updates the role of an account and saves associated role-specific details.
// @Summary Update account role and role-specific details
// @Description Update the role of an account and save associated role-specific details.
// @Tags account
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token (e.g., 'Bearer <token>')"
// @Param payload body dto.UpdateRolePayloadDTO true "Payload containing role and role-specific details"
// @Success 200 {object} map[string]interface{} "Account role updated successfully"
// @Failure 400 {object} response.GeneralError "Invalid payload"
// @Failure 404 {object} response.GeneralError "Account not found"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/account/role [put]
func (h *accountHandler) UpdateAccountRole(ctx *gin.Context) {
	// Retrieve the token from the context
	token, exists := ctx.Get("token")
	if !exists {
		httpresponse.Error(ctx, http.StatusUnauthorized, "Token not found", nil)
		return
	}

	// Validate the token and fetch user details
	accountDTO, err := h.authUCase.ValidateToken(token.(string))
	if err != nil {
		logger.GetLogger().Errorf("Failed to validate token: %v", err)
		httpresponse.Error(ctx, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	var req dto.UpdateRolePayloadDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid payload", err)
		return
	}

	// Validate the role
	role := constants.AccountRole(req.Role)
	if !role.IsValid() {
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid role provided", nil)
		return
	}

	// Fetch the account by ID
	account, err := h.accountUCase.FindAccountByID(accountDTO.ID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to fetch account: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to fetch account", err)
		return
	}
	if account == nil {
		httpresponse.Error(ctx, http.StatusNotFound, "Account not found", nil)
		return
	}

	// Update the role
	account.Role = req.Role
	if err := h.accountUCase.UpdateAccount(account); err != nil {
		logger.GetLogger().Errorf("Failed to update account role: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to update account role", err)
		return
	}

	// Save role-specific details
	if err := h.authUCase.UpdateRoleDetail(accountDTO.ID, role, &req.RoleDetails); err != nil {
		logger.GetLogger().Errorf("Failed to save role-specific details: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to save role-specific details", err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Account role updated successfully"})
}
