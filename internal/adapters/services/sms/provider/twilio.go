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

// TwilioProvider handles SMS through Twilio
type TwilioProvider struct {
	client *client.TwilioClient
	config conf.TwilioConfiguration
}

func NewTwilioProvider(config conf.TwilioConfiguration) SMSProvider {
	return &TwilioProvider{
		client: client.NewTwilioClient(config.TwilioAccountSID, config.TwilioAuthToken, config.TwilioBaseURL),
		config: config,
	}
}

func (t *TwilioProvider) SendOTP(ctx context.Context, tenantName, receiver, otp string, ttl time.Duration) error {
	logger.GetLogger().Infof("Sending SMS to %s via Twilio", receiver)

	message := common.GetOTPMessage(tenantName, otp, ttl)
	resp, err := t.client.SendSMS(tenantName, t.config.TwilioFrom, receiver, message)
	if err != nil {
		return fmt.Errorf("failed to send SMS via Twilio: %w", err)
	}

	logger.GetLogger().Infof("SMS sent successfully via Twilio: %+v", resp)
	return nil
}

func (t *TwilioProvider) RefreshToken(ctx context.Context) error {
	// Twilio doesn't require token refresh - using API key authentication
	return nil
}

func (t *TwilioProvider) GetChannelType() string {
	return constants.ChannelSMS
}

func (t *TwilioProvider) HealthCheck(ctx context.Context) error {
	// Implement Twilio health check if available
	return nil
}
