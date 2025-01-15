package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

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

// DataUploadWebhook handles notifications when a user uploads data successfully.
// @Summary Notify about successful data upload
// @Description This webhook receives raw payload data when a user successfully uploads data.
// @Tags notifications
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token (e.g., 'Bearer <token>')"
// @Success 201 {object} map[string]interface{} "Notification received successfully"
// @Failure 400 {object} response.GeneralError "Invalid payload"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/notifications/data-upload [post]
func (h *notificationHandler) DataUploadWebhook(ctx *gin.Context) {
	// Retrieve the authenticated account from context
	accountDTO, ok := ctx.Get("account")
	if !ok {
		httpresponse.Error(ctx, http.StatusUnauthorized, "Unauthorized access: account information missing", nil)
		return
	}

	// Read the raw request body
	body, err := ctx.GetRawData()
	if err != nil {
		logger.GetLogger().Errorf("Failed to read request body: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to read request body", err)
		return
	}

	// TODO: Implement data upload webhook processing logic here
	fmt.Println(accountDTO)
	fmt.Println(body)

	// Return success response
	ctx.JSON(http.StatusCreated, gin.H{"message": "Notification received successfully"})
}
