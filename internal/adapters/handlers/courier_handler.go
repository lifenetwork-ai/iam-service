package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
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
