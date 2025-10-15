package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms"
	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
	httpresponse "github.com/lifenetwork-ai/iam-service/packages/http/response"
)

type SmsTokenHandler struct {
	uc            interfaces.SmsTokenUseCase
	smsService    *sms.SMSService
	zaloTokenRepo domainrepo.ZaloTokenRepository
}

func NewSmsTokenHandler(uc interfaces.SmsTokenUseCase, smsService *sms.SMSService, zaloTokenRepo domainrepo.ZaloTokenRepository) *SmsTokenHandler {
	return &SmsTokenHandler{uc: uc, smsService: smsService, zaloTokenRepo: zaloTokenRepo}
}

// @Summary Get Zalo token
// @Description Get Zalo token
// @Security BasicAuth
// @Tags sms
// @Accept json
// @Produce json
// @Success 200 {object} dto.ZaloTokenResponseDTO "Zalo token"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/admin/sms/zalo/token [get]
func (h *SmsTokenHandler) GetZaloToken(c *gin.Context) {
	token, derr := h.uc.GetZaloToken(c.Request.Context())
	if derr != nil {
		httpresponse.Error(c, http.StatusInternalServerError, derr.Code, derr.Message, nil)
		return
	}

	resp := dto.ZaloTokenResponseDTO{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.ExpiresAt.Local().Format(time.RFC3339),
		UpdatedAt:    token.UpdatedAt.Local().Format(time.RFC3339),
	}
	httpresponse.Success(c, http.StatusOK, resp)
}

// GetZaloHealth checks Zalo provider health
// @Summary Zalo provider health check
// @Description Perform a health check against the Zalo API using current token
// @Security BasicAuth
// @Tags sms
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/admin/sms/zalo/health [get]
func (h *SmsTokenHandler) GetZaloHealth(c *gin.Context) {
	if err := h.uc.ZaloHealthCheck(c.Request.Context()); err != nil {
		httpresponse.Error(c, http.StatusInternalServerError, "MSG_ZALO_HEALTH_FAIL", err.Error(), nil)
		return
	}

	httpresponse.Success(c, http.StatusOK, map[string]string{"status": "healthy"})
}

// RefreshZaloToken refreshes the Zalo token and returns the new token
// @Summary Refresh Zalo token
// @Description Refresh the Zalo access token using the refresh token and return the new token details
// @Security BasicAuth
// @Tags sms
// @Accept json
// @Produce json
// @Param request body dto.RefreshZaloTokenRequestDTO true "Refresh Zalo token request"
// @Success 200 {object} dto.RefreshZaloTokenResponseDTO "Refreshed Zalo token"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/admin/sms/zalo/token/refresh [post]
func (h *SmsTokenHandler) RefreshZaloToken(c *gin.Context) {
	var req dto.RefreshZaloTokenRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresponse.Error(c, http.StatusBadRequest, "MSG_INVALID_REQUEST", "Invalid request body", []map[string]string{{"field": "body", "error": err.Error()}})
		return
	}

	if derr := h.uc.RefreshZaloToken(c.Request.Context(), req.RefreshToken); derr != nil {
		httpresponse.Error(c, http.StatusInternalServerError, derr.Code, derr.Message, nil)
		return
	}

	token, derr := h.uc.GetZaloToken(c.Request.Context())
	if derr != nil {
		httpresponse.Error(c, http.StatusInternalServerError, derr.Code, derr.Message, nil)
		return
	}

	resp := dto.ZaloTokenResponseDTO{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.ExpiresAt.Local().Format(time.RFC3339),
		UpdatedAt:    token.UpdatedAt.Local().Format(time.RFC3339),
	}
	httpresponse.Success(c, http.StatusOK, resp)
}
