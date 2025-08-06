package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/constants"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms/client"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms/common"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

// WhatsAppProvider handles messages through WhatsApp Business API
type WhatsAppProvider struct {
	client *client.WhatsAppClient
	config conf.WhatsappConfiguration
}

func NewWhatsAppProvider(config conf.WhatsappConfiguration) *WhatsAppProvider {
	return &WhatsAppProvider{
		client: client.NewWhatsAppClient(config.WhatsappAccessToken, config.WhatsappPhoneID, config.WhatsappBaseURL),
		config: config,
	}
}

func (w *WhatsAppProvider) SendOTP(ctx context.Context, tenantName, receiver, otp string, ttl time.Duration) error {
	logger.GetLogger().Infof("Sending OTP to %s via WhatsApp", receiver)

	message := common.GetOTPMessage(tenantName, otp, ttl)
	resp, err := w.client.SendMessage(tenantName, receiver, message)
	if err != nil {
		return fmt.Errorf("failed to send message via WhatsApp: %w", err)
	}

	logger.GetLogger().Infof("WhatsApp message sent successfully: %+v", resp)
	return nil
}

func (w *WhatsAppProvider) RefreshToken(ctx context.Context) error {
	// TODO: Implement WhatsApp token refresh when needed
	// For now, WhatsApp uses long-lived tokens
	return nil
}

func (w *WhatsAppProvider) GetChannelType() string {
	return constants.ChannelWhatsApp
}

func (w *WhatsAppProvider) HealthCheck(ctx context.Context) error {
	// Implement WhatsApp health check if available
	return nil
}
