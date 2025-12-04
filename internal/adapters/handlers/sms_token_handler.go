package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms"
	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	"github.com/lifenetwork-ai/iam-service/internal/delivery/http/middleware"
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
// @Description Get Zalo token for a specific tenant
// @Security BasicAuth
// @Tags sms
// @Accept json
// @Produce json
// @Param X-Tenant-Id header string true "Tenant ID"
// @Success 200 {object} dto.ZaloTokenResponseDTO "Zalo token"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/admin/sms/zalo/token [get]
func (h *SmsTokenHandler) GetZaloToken(c *gin.Context) {
	tenant, err := middleware.GetTenantFromContext(c)
	if err != nil {
		httpresponse.Error(c, http.StatusBadRequest, "MSG_TENANT_NOT_FOUND", "Tenant not found in context", nil)
		return
	}

	token, derr := h.uc.GetEncryptedZaloToken(c.Request.Context(), tenant.ID)
	if derr != nil {
		httpresponse.Error(c, http.StatusInternalServerError, derr.Code, derr.Message, nil)
		return
	}

	resp := dto.ZaloTokenResponseDTO{
		TenantID:    token.TenantID.String(),
		AppID:       token.AppID,
		SecretKey:   token.SecretKey,
		AccessToken: token.AccessToken,
		// RefreshToken: token.RefreshToken,
		ExpiresAt: token.ExpiresAt.Local().Format(time.RFC3339),
		UpdatedAt: token.UpdatedAt.Local().Format(time.RFC3339),
	}
	httpresponse.Success(c, http.StatusOK, resp)
}

// GetZaloHealth checks Zalo provider health
// @Summary Zalo provider health check
// @Description Perform a health check against the Zalo API using tenant's token
// @Security BasicAuth
// @Tags sms
// @Accept json
// @Produce json
// @Param X-Tenant-Id header string true "Tenant ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/admin/sms/zalo/health [get]
func (h *SmsTokenHandler) GetZaloHealth(c *gin.Context) {
	tenant, err := middleware.GetTenantFromContext(c)
	if err != nil {
		httpresponse.Error(c, http.StatusBadRequest, "MSG_TENANT_NOT_FOUND", "Tenant not found in context", nil)
		return
	}

	if derr := h.uc.ZaloHealthCheck(c.Request.Context(), tenant.ID); derr != nil {
		httpresponse.Error(c, http.StatusInternalServerError, "MSG_ZALO_HEALTH_FAIL", derr.Error(), nil)
		return
	}

	httpresponse.Success(c, http.StatusOK, map[string]string{"status": "healthy"})
}

// CreateOrUpdateZaloToken creates or updates Zalo token configuration for a tenant
// @Summary Create or update Zalo token
// @Description Create or update Zalo token configuration including app credentials
// @Security BasicAuth
// @Tags sms
// @Accept json
// @Produce json
// @Param X-Tenant-Id header string true "Tenant ID"
// @Param request body dto.CreateZaloTokenRequestDTO true "Create Zalo token request"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/admin/sms/zalo/token [post]
func (h *SmsTokenHandler) CreateOrUpdateZaloToken(c *gin.Context) {
	tenant, err := middleware.GetTenantFromContext(c)
	if err != nil {
		httpresponse.Error(c, http.StatusBadRequest, "MSG_TENANT_NOT_FOUND", "Tenant not found in context", nil)
		return
	}

	var req dto.CreateZaloTokenRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresponse.Error(c, http.StatusBadRequest, "MSG_INVALID_REQUEST", "Invalid request body", []map[string]string{{"field": "body", "error": err.Error()}})
		return
	}

	if derr := h.uc.CreateOrUpdateZaloToken(c.Request.Context(), tenant.ID, req.AppID, req.SecretKey, req.RefreshToken, req.AccessToken, req.OtpTemplateID); derr != nil {
		httpresponse.Error(c, http.StatusInternalServerError, derr.Code, derr.Message, nil)
		return
	}

	httpresponse.Success(c, http.StatusOK, map[string]string{"message": "Zalo token created/updated successfully"})
}

// RefreshZaloToken refreshes the Zalo token and returns the new token
// @Summary Refresh Zalo token
// @Description Refresh the Zalo access token using the refresh token and return the new token details
// @Security BasicAuth
// @Tags sms
// @Accept json
// @Produce json
// @Param X-Tenant-Id header string true "Tenant ID"
// @Param request body dto.RefreshZaloTokenRequestDTO true "Refresh Zalo token request"
// @Success 200 {object} dto.RefreshZaloTokenResponseDTO "Refreshed Zalo token"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/admin/sms/zalo/token/refresh [post]
func (h *SmsTokenHandler) RefreshZaloToken(c *gin.Context) {
	tenant, err := middleware.GetTenantFromContext(c)
	if err != nil {
		httpresponse.Error(c, http.StatusBadRequest, "MSG_TENANT_NOT_FOUND", "Tenant not found in context", nil)
		return
	}

	var req dto.RefreshZaloTokenRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresponse.Error(c, http.StatusBadRequest, "MSG_INVALID_REQUEST", "Invalid request body", []map[string]string{{"field": "body", "error": err.Error()}})
		return
	}

	if derr := h.uc.RefreshZaloToken(c.Request.Context(), tenant.ID, req.RefreshToken); derr != nil {
		httpresponse.Error(c, http.StatusInternalServerError, derr.Code, derr.Message, nil)
		return
	}

	token, derr := h.uc.GetZaloToken(c.Request.Context(), tenant.ID)
	if derr != nil {
		httpresponse.Error(c, http.StatusInternalServerError, derr.Code, derr.Message, nil)
		return
	}

	resp := dto.RefreshZaloTokenResponseDTO{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.ExpiresAt.Local().Format(time.RFC3339),
		UpdatedAt:    token.UpdatedAt.Local().Format(time.RFC3339),
	}
	httpresponse.Success(c, http.StatusOK, resp)
}

// DeleteZaloToken deletes the Zalo token configuration for a tenant
// @Summary Delete Zalo token
// @Description Delete Zalo token configuration for a specific tenant
// @Security BasicAuth
// @Tags sms
// @Accept json
// @Produce json
// @Param X-Tenant-Id header string true "Tenant ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/admin/sms/zalo/token [delete]
func (h *SmsTokenHandler) DeleteZaloToken(c *gin.Context) {
	tenant, err := middleware.GetTenantFromContext(c)
	if err != nil {
		httpresponse.Error(c, http.StatusBadRequest, "MSG_TENANT_NOT_FOUND", "Tenant not found in context", nil)
		return
	}

	if derr := h.uc.DeleteZaloToken(c.Request.Context(), tenant.ID); derr != nil {
		httpresponse.Error(c, http.StatusInternalServerError, derr.Code, derr.Message, nil)
		return
	}

	httpresponse.Success(c, http.StatusOK, map[string]string{"message": "Zalo token deleted successfully"})
}
