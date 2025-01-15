package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/human-network-auth/internal/dto"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
	httpresponse "github.com/genefriendway/human-network-auth/pkg/http/response"
	"github.com/genefriendway/human-network-auth/pkg/logger"
)

type notificationHandler struct {
	authUCase interfaces.AuthUCase
}

func NewNotificationHandler(authUCase interfaces.AuthUCase) *notificationHandler {
	return &notificationHandler{authUCase: authUCase}
}

// DataUploadNotification handles notifications when a user uploads data successfully.
// @Summary Notify about successful data upload
// @Description This endpoint receives notifications when a user successfully uploads data.
// @Tags notifications
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token (e.g., 'Bearer <token>')"
// @Param request body dto.DataUploadNotificationPayloadDTO true "Payload containing data upload details"
// @Success 201 {object} map[string]interface{} "Notification received successfully"
// @Failure 400 {object} response.GeneralError "Invalid payload"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/notifications/data-upload [post]
func (h *notificationHandler) DataUploadNotification(ctx *gin.Context) {
	// Retrieve the authenticated account from context
	accountDTO, ok := ctx.Get("account")
	if !ok {
		httpresponse.Error(ctx, http.StatusUnauthorized, "Unauthorized access: account information missing", nil)
		return
	}

	// Parse the request payload
	var payload dto.DataUploadNotificationPayloadDTO
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid payload", err)
		return
	}

	// Process the notification
	// TODO: Implement notification processing
	fmt.Println("Account ID:", accountDTO.(*dto.AccountDTO).ID)

	// Return success response
	ctx.JSON(http.StatusCreated, gin.H{"message": "Notification received successfully"})
}
