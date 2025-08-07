package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
	httpresponse "github.com/lifenetwork-ai/iam-service/packages/http/response"
)

type smsTokenHandler struct {
	ucase interfaces.SmsTokenUseCase
}

func NewSmsTokenHandler(ucase interfaces.SmsTokenUseCase) *smsTokenHandler {
	return &smsTokenHandler{ucase: ucase}
}

// @Summary Get Zalo token
// @Description Get Zalo token
// @Security BasicAuth
// @Tags Zalo
// @Accept json
// @Produce json
// @Success 200 {object} dto.ZaloTokenResponseDTO
// @Router /zalo/token [get]
func (h *smsTokenHandler) GetZaloToken(ctx *gin.Context) {
	token, err := h.ucase.GetZaloToken(ctx)
	if err != nil {
		handleDomainError(ctx, err)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, dto.ZaloTokenResponseDTO{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	})
}

// @Summary Set Zalo token
// @Description Set Zalo token
// @Security BasicAuth
// @Tags Zalo
// @Accept json
// @Produce json
// @Param token body dto.CreateZaloTokenRequestDTO true "Zalo token"
// @Success 200 {object} dto.ZaloTokenResponseDTO
// @Router /zalo/token [post]
func (h *smsTokenHandler) SetZaloToken(ctx *gin.Context) {
	var req dto.CreateZaloTokenRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		handleDomainError(ctx, domainerrors.NewValidationError("MSG_INVALID_REQUEST", "Invalid request", err))
		return
	}

	err := h.ucase.SetZaloToken(ctx, req.AccessToken, req.RefreshToken)
	if err != nil {
		handleDomainError(ctx, err)
		return
	}

	httpresponse.Success(ctx, http.StatusOK, "Zalo token set successfully")
}
