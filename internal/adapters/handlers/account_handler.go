package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/dto"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
	httpresponse "github.com/genefriendway/human-network-auth/pkg/http/response"
	"github.com/genefriendway/human-network-auth/pkg/logger"
	"github.com/genefriendway/human-network-auth/pkg/utils"
)

type accountHandler struct {
	iamUCase     interfaces.IAMUCase
	accountUCase interfaces.AccountUCase
	authUCase    interfaces.AuthUCase
}

func NewAccountHandler(
	iamUCase interfaces.IAMUCase,
	accountUCase interfaces.AccountUCase,
	authUCase interfaces.AuthUCase,
) *accountHandler {
	return &accountHandler{
		iamUCase:     iamUCase,
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
// @Failure 403 {object} response.GeneralError "Insufficient permissions"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/account/me [get]
func (h *accountHandler) GetCurrentUser(ctx *gin.Context) {
	// Retrieve the authenticated account from the context
	accountDTO, exists := ctx.Get("account")
	if !exists {
		httpresponse.Error(ctx, http.StatusUnauthorized, "Unauthorized access: account information missing", nil)
		return
	}

	// Fetch role-specific details based on the authenticated account
	detail, err := h.accountUCase.FindDetailByAccountID(
		accountDTO.(*dto.AccountDTO),
		constants.AccountRole(accountDTO.(*dto.AccountDTO).Role),
	)
	if err != nil {
		logger.GetLogger().Errorf("Failed to fetch user details for account ID [%s]: %v", accountDTO.(*dto.AccountDTO).ID, err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to fetch user details", err)
		return
	}

	// Respond with user details
	ctx.JSON(http.StatusOK, detail)
}

// GetActiveValidators retrieves the list of active validators.
// @Summary Get Active Validators
// @Description Fetches a list of active validators. Optionally, a comma-separated list of validator IDs can be provided to filter the results.
// @Tags validators
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token (e.g., 'Bearer <token>')"
// @Param validator_ids query string false "Comma-separated list of validator IDs to filter results (e.g., 'id1,id2,id3')"
// @Success 200 {array} dto.AccountDetailDTO "List of active validators"
// @Failure 400 {object} response.GeneralError "Bad request"
// @Failure 401 {object} response.GeneralError "Unauthorized"
// @Failure 403 {object} response.GeneralError "Insufficient permissions"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/validators/active [get]
func (h *accountHandler) GetActiveValidators(ctx *gin.Context) {
	// Retrieve the authenticated account from the context
	_, exists := ctx.Get("account")
	if !exists {
		httpresponse.Error(ctx, http.StatusUnauthorized, "Unauthorized access: account information missing", nil)
		return
	}

	// Parse the optional query parameter for validator IDs
	validatorIDsParam := ctx.Query("validator_ids")
	var validatorIDs []string
	if validatorIDsParam != "" {
		validatorIDs = strings.Split(validatorIDsParam, ",")
		for i := range validatorIDs {
			validatorIDs[i] = strings.TrimSpace(validatorIDs[i])
		}
	}

	// Fetch active validators, optionally filtered by IDs
	validators, err := h.accountUCase.GetActiveValidators(validatorIDs)
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
// @Failure 401 {object} response.GeneralError "Unauthorized"
// @Failure 403 {object} response.GeneralError "Insufficient permissions"
// @Failure 404 {object} response.GeneralError "Account not found"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/account/role [put]
func (h *accountHandler) UpdateAccountRole(ctx *gin.Context) {
	// Retrieve the authenticated account from context
	accountDTO, ok := ctx.Get("account")
	if !ok {
		httpresponse.Error(ctx, http.StatusUnauthorized, "Unauthorized access: account information missing", nil)
		return
	}

	// Parse and validate the request payload
	var req dto.UpdateRolePayloadDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Errorf("Payload binding failed: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid payload format", err)
		return
	}

	// Validate the new role
	role := constants.AccountRole(req.Role)
	if !role.IsValid() {
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid role specified", nil)
		return
	}

	// Fetch the account by ID
	account, err := h.accountUCase.FindAccountByID(accountDTO.(*dto.AccountDTO).ID)
	if err != nil {
		logger.GetLogger().Errorf("Error fetching account [ID: %s]: %v", accountDTO.(*dto.AccountDTO).ID, err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Error fetching account", err)
		return
	}
	if account == nil {
		httpresponse.Error(ctx, http.StatusNotFound, "Account not found", nil)
		return
	}

	// Update the account's role
	account.Role = req.Role
	if err := h.accountUCase.UpdateAccount(account); err != nil {
		logger.GetLogger().Errorf("Error updating account role [ID: %s, Role: %s]: %v", account.ID, req.Role, err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to update account role", err)
		return
	}

	// Update role-specific details
	if err := h.authUCase.UpdateRoleDetail(accountDTO.(*dto.AccountDTO).ID, role, &req.RoleDetails); err != nil {
		logger.GetLogger().Errorf(
			"Error updating role-specific details [ID: %s, Role: %s]: %v", accountDTO.(*dto.AccountDTO).ID, req.Role, err,
		)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to save role-specific details", err)
		return
	}

	// Synchronize policies for the new role
	err = h.syncAccountPolicies(account.ID, role)
	if err != nil {
		logger.GetLogger().Errorf("Failed to synchronize account policies: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to synchronize account policies", err)
		return
	}

	// Respond with success
	logger.GetLogger().Infof("Successfully updated role [AccountID: %s, NewRole: %s]", account.ID, req.Role)
	ctx.JSON(http.StatusOK, gin.H{"message": "Account role updated successfully"})
}

// syncAccountPolicies synchronizes the policies assigned to an account based on its role.
func (h *accountHandler) syncAccountPolicies(accountID string, role constants.AccountRole) error {
	// Fetch predefined policies for the role
	predefinedPolicy, err := h.iamUCase.GetPolicyByRole(role)
	if err != nil {
		return err
	}

	// Remove existing policies for the account
	if err := h.iamUCase.RemovePoliciesFromAccount(accountID); err != nil {
		return err
	}

	// Assign the new policy to the account
	if err := h.iamUCase.AssignPolicyToAccount(accountID, predefinedPolicy.ID); err != nil {
		return err
	}

	return nil
}

// UpdateAPIKey creates or updates an API key for a specific account.
// @Summary Create or update API key
// @Description Generate or update the API key for an account.
// @Tags account
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token (e.g., 'Bearer <token>')"
// @Success 200 {object} map[string]interface{} "API key updated successfully"
// @Failure 401 {object} response.GeneralError "Unauthorized"
// @Failure 403 {object} response.GeneralError "Forbidden"
// @Failure 404 {object} response.GeneralError "Account not found"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/account/api-key [put]
func (h *accountHandler) UpdateAPIKey(ctx *gin.Context) {
	// Retrieve the authenticated account from the context
	accountDTO, exists := ctx.Get("account")
	if !exists {
		httpresponse.Error(ctx, http.StatusUnauthorized, "Unauthorized access: account information missing", nil)
		return
	}

	// Fetch the account from the database
	account, err := h.accountUCase.FindAccountByID(accountDTO.(*dto.AccountDTO).ID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to fetch account: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Error fetching account", err)
		return
	}
	if account == nil {
		httpresponse.Error(ctx, http.StatusNotFound, "Account not found", nil)
		return
	}

	// Generate a new API key
	newAPIKey, err := utils.GenerateAPIKey()
	if err != nil {
		logger.GetLogger().Errorf("Failed to generate API key: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Error generating API key", err)
		return
	}

	// Update the account with the new API key
	account.APIKey = &newAPIKey
	if err := h.accountUCase.UpdateAccount(account); err != nil {
		logger.GetLogger().Errorf("Failed to update account with API key: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Error updating account with API key", err)
		return
	}

	// Respond with success and return the new API key
	ctx.JSON(http.StatusOK, gin.H{
		"message": "API key updated successfully",
		"api_key": newAPIKey,
	})
}
