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
	geneticaClient *client.SpeedSMSClient
	lifeClient     *client.SpeedSMSClient
}

func NewSpeedSMSProvider(conf conf.SpeedSMSConfiguration) SMSProvider {
	return &SpeedSMSProvider{
		geneticaClient: client.NewSpeedSMSClient(conf.GeneticaSpeedSMSAccessToken),
		lifeClient:     client.NewSpeedSMSClient(conf.LifeSpeedSMSAccessToken),
	}
}

func (s *SpeedSMSProvider) SendOTP(ctx context.Context, tenantName, receiver, otp string, ttl time.Duration) error {
	var brandname client.SpeedSMSBrandname
	var smsClient *client.SpeedSMSClient

	// Select client and brandname based on tenant
	if tenantName == constants.TenantLifeAI {
		brandname = constants.LifeBrandname
		smsClient = s.lifeClient
	} else {
		brandname = constants.GeneticaBrandname
		smsClient = s.geneticaClient
	}

	resp, err := smsClient.SendOTP(receiver, otp, brandname)
	if err != nil {
		return fmt.Errorf("failed to send SMS via SpeedSMS for tenant %s: %w", tenantName, err)
	}
	logger.GetLogger().Infof("SMS sent successfully via SpeedSMS for tenant %s: %+v", tenantName, resp)
	return nil
}

func (s *SpeedSMSProvider) GetChannelType() string {
	return constants.ChannelSpeedSMS
}

func (s *SpeedSMSProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func (s *SpeedSMSProvider) RefreshToken(ctx context.Context, refreshToken string) error {
	return nil
}
