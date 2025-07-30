package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	"github.com/lifenetwork-ai/iam-service/internal/delivery/http/middleware"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
	httpresponse "github.com/lifenetwork-ai/iam-service/packages/http/response"
)

type courierHandler struct {
	ucase interfaces.CourierUseCase
}

func NewCourierHandler(ucase interfaces.CourierUseCase) *courierHandler {
	return &courierHandler{
		ucase: ucase,
	}
}

// ReceiveCourierMessageHandler receives a raw courier message and stores it in the courier queue.
// @Summary Receive courier message (from webhook or sender)
// @Description Receive courier content and enqueue it for delivery
// @Tags courier
// @Accept json
// @Produce json
// @Param payload body dto.CourierWebhookRequestDTO true "Courier message payload"
// @Success 200 {object} response.SuccessResponse "Courier message enqueued successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/courier/messages [post]
func (h *courierHandler) ReceiveCourierMessageHandler(ctx *gin.Context) {
	var req dto.CourierWebhookRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		httpresponse.Error(ctx, http.StatusBadRequest, "MSG_INVALID_PAYLOAD", "Invalid request payload", err)
		return
	}

	if req.To == "" || req.Body == "" {
		httpresponse.Error(ctx, http.StatusBadRequest, "MSG_INVALID_FIELDS", "To and Body are required", nil)
		return
	}

	if err := h.ucase.ReceiveOTP(ctx, req.To, req.Body); err != nil {
		handleDomainError(ctx, err)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, gin.H{"message": "OTP received successfully"})
}

// GetAvailableChannelsHandler returns available OTP delivery channels based on tenant and receiver
// @Summary Get available delivery channels
// @Description Returns available delivery channels (SMS, WhatsApp, Zalo) based on receiver and tenant
// @Param X-Tenant-Id header string true "Tenant ID"
// @Param receiver query string true "Receiver identifier"
// @Tags courier
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=[]string} "List of available channels"
// @Failure 400 {object} response.ErrorResponse "Invalid receiver"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Router /api/v1/courier/available-channels [get]
func (h *courierHandler) GetAvailableChannelsHandler(ctx *gin.Context) {
	// Get tenant from context
	tenant, err := middleware.GetTenantFromContext(ctx)
	if err != nil {
		httpresponse.Error(ctx, http.StatusBadRequest, "MSG_INVALID_TENANT", "Invalid tenant", err)
		return
	}

	var req dto.CourierGetAvailableChannelsRequestDTO
	if err := ctx.ShouldBindQuery(&req); err != nil {
		httpresponse.Error(ctx, http.StatusBadRequest, "MSG_INVALID_PAYLOAD", "Invalid request payload", err)
		return
	}

	channels := h.ucase.GetAvailableChannels(ctx, tenant.Name, req.Receiver)
	httpresponse.Success(ctx, http.StatusOK, channels)
}

// ChooseChannelHandler chooses a channel for a receiver
// @Summary Choose a channel for a receiver
// @Description Chooses a channel for a receiver (SMS, WhatsApp, Zalo)
// @Param X-Tenant-Id header string true "Tenant ID"
// @Tags courier
// @Accept json
// @Produce json
// @Param payload body dto.CourierChooseChannelRequestDTO true "Channel and receiver"
// @Success 200 {object} response.SuccessResponse "Channel chosen successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/courier/choose-channel [post]
func (h *courierHandler) ChooseChannelHandler(ctx *gin.Context) {
	var req dto.CourierChooseChannelRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		httpresponse.Error(ctx, http.StatusBadRequest, "MSG_INVALID_PAYLOAD", "Invalid request payload", err)
		return
	}

	if req.Channel == "" || req.Receiver == "" {
		httpresponse.Error(ctx, http.StatusBadRequest, "MSG_INVALID_FIELDS", "Channel and Receiver are required", nil)
		return
	}

	tenant, err := middleware.GetTenantFromContext(ctx)
	if err != nil {
		httpresponse.Error(ctx, http.StatusBadRequest, "MSG_INVALID_TENANT", "Invalid tenant", err)
		return
	}

	usecaseErr := h.ucase.ChooseChannel(ctx, tenant.Name, req.Receiver, req.Channel)
	if usecaseErr != nil {
		handleDomainError(ctx, usecaseErr)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, "Channel chosen successfully")
}
