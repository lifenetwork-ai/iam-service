package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/constants"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms/common"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

// WebhookProvider handles messages through webhook (fallback)
type WebhookProvider struct{}

func NewWebhookProvider() SMSProvider {
	return &WebhookProvider{}
}

func (w *WebhookProvider) SendOTP(ctx context.Context, tenantName, receiver, otp string, ttl time.Duration, _ string) error {
	logger.GetLogger().Infof("Sending OTP to %s via webhook", receiver)

	url := conf.GetMockWebhookURL()
	if url == "" {
		return errors.New("MOCK_WEBHOOK_URL is not set")
	}

	type webhookPayload struct {
		Tenant  string `json:"tenant"`
		To      string `json:"to"`
		Message string `json:"message"`
		OTP     string `json:"otp"`
		TTL     int64  `json:"ttl_seconds"`
	}

	message := common.GetOTPMessage(tenantName, otp, ttl)
	payload := webhookPayload{
		Tenant:  tenantName,
		To:      receiver,
		Message: message,
		OTP:     otp,
		TTL:     int64(ttl.Seconds()),
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}
	req.Header.Set(constants.HeaderKeyContentType, constants.HeaderContentTypeJson)

	client := &http.Client{Timeout: constants.WebhookTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-2xx status: %s", resp.Status)
	}

	logger.GetLogger().Infof("Webhook sent successfully to %s", receiver)
	return nil
}

func (w *WebhookProvider) RefreshToken(ctx context.Context, refreshToken string) error {
	// Webhook doesn't require token refresh
	return nil
}

func (w *WebhookProvider) GetChannelType() string {
	return constants.DefaultSMSChannel
}

func (w *WebhookProvider) HealthCheck(ctx context.Context) error {
	// Check if webhook URL is configured
	url := conf.GetMockWebhookURL()
	if url == "" {
		return errors.New("MOCK_WEBHOOK_URL is not set")
	}
	return nil
}
