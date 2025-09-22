package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/constants"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms/client"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

type SpeedSMSProvider struct {
	client *client.SpeedSMSClient
}

func NewSpeedSMSProvider(conf conf.SpeedSMSConfiguration) SMSProvider {
	return &SpeedSMSProvider{
		client: client.NewSpeedSMSClient(conf.SpeedSMSAccessToken),
	}
}

func (s *SpeedSMSProvider) SendOTP(ctx context.Context, tenantName, receiver, otp string, ttl time.Duration) error {
	resp, err := s.client.SendVerificationSMS(receiver, otp, "LIFE AI")
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}
	logger.GetLogger().Infof("SMS sent successfully via SpeedSMS: %+v", resp)
	return nil
}

func (s *SpeedSMSProvider) GetChannelType() string {
	return constants.ChannelSMS
}

func (s *SpeedSMSProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func (s *SpeedSMSProvider) RefreshToken(ctx context.Context, refreshToken string) error {
	return nil
}
