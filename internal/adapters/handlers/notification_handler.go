package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/human-network-auth/internal/dto"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
	httpresponse "github.com/genefriendway/human-network-auth/pkg/http/response"
	"github.com/genefriendway/human-network-auth/pkg/logger"
	"github.com/genefriendway/human-network-auth/pkg/utils"
)

type notificationHandler struct {
	authUCase       interfaces.AuthUCase
	accountUCase    interfaces.AccountUCase
	dataAccessUCase interfaces.DataAccessUCase
	fileInfoUCase   interfaces.FileInfoUCase
}

func NewNotificationHandler(
	authUCase interfaces.AuthUCase,
	accountUCase interfaces.AccountUCase,
	dataAccessUCase interfaces.DataAccessUCase,
	fileInfoUCase interfaces.FileInfoUCase,
) *notificationHandler {
	return &notificationHandler{
		authUCase:       authUCase,
		accountUCase:    accountUCase,
		dataAccessUCase: dataAccessUCase,
		fileInfoUCase:   fileInfoUCase,
	}
}

// DataUploadWebhook handles notifications when a user uploads data successfully.
// @Summary Notify about successful data upload
// @Description This webhook receives raw payload data when a user successfully uploads data.
// @Tags notifications
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token (e.g., 'Bearer <token>')"
// @Param payload body dto.FileInfoPayloadDTO true "Payload containing file information"
// @Success 201 {object} map[string]interface{} "Notification received successfully"
// @Failure 400 {object} response.GeneralError "Invalid payload"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/notifications/data-upload [post]
func (h *notificationHandler) DataUploadWebhook(ctx *gin.Context) {
	fmt.Println("DataUploadWebhook...")
	// Retrieve the authenticated account from context
	accountDTO, ok := ctx.Get("account")
	if !ok {
		httpresponse.Error(ctx, http.StatusUnauthorized, "Unauthorized access: account information missing", nil)
		return
	}

	// Parse the JSON payload into a struct
	payload := &dto.FileInfoPayloadDTO{}
	if err := ctx.ShouldBindJSON(payload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid payload format", err)
		return
	}

	validators, err := h.accountUCase.GetActiveValidators([]string{})
	if err != nil {
		logger.GetLogger().Errorf("Failed to get active validators: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to get active validators", err)
		return
	}

	// Select a random subset of validators
	subsetSize := 3
	selectedValidators, err := utils.SelectRandomSubset(validators, subsetSize)
	if err != nil {
		logger.GetLogger().Errorf("Failed to select random validators: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to select random validators", err)
		return
	}
	logger.GetLogger().Infof("Selected validators: %v", selectedValidators)

	// Convert selected validators to DTOs
	requesterAccounts := make([]dto.AccountDTO, len(selectedValidators))
	for i, validator := range selectedValidators {
		requesterAccounts[i] = validator.Account
	}

	// Create a data access request for the selected validators
	dataAccessPayload := dto.DataAccessRequestPayloadDTO{
		RequestAccountID: accountDTO.(*dto.AccountDTO).ID,
		ReasonForRequest: "Access data for validation",
		FileID:           payload.ID,
	}

	if err := h.dataAccessUCase.CreateRequest(dataAccessPayload, requesterAccounts); err != nil {
		logger.GetLogger().Errorf("Failed to create data access request: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to create data access request", err)
		return
	}

	_ = h.fileInfoUCase.CreateFileInfo(*payload)

	// TODO: map the data access request

	// Return success response
	ctx.JSON(http.StatusCreated, gin.H{"message": "Notification received successfully"})
}
