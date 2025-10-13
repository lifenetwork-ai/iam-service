package sms

import (
	"context"
	"fmt"
	"time"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/constants"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms/provider"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

// SMSProviderFactory manages and creates SMS providers
type SMSProviderFactory struct {
	providers map[string]provider.SMSProvider
}

// NewSMSProviderFactory creates a new factory with all configured providers
// Don't return error because we want to continue initializing the service even if some providers are not configured
func NewSMSProviderFactory(config *conf.SmsConfiguration, zaloTokenRepo domainrepo.ZaloTokenRepository) (*SMSProviderFactory, error) {
	factory := &SMSProviderFactory{
		providers: make(map[string]provider.SMSProvider),
	}

	// Initialize Twilio provider
	if config.Twilio.TwilioAccountSID != "" {
		factory.providers[constants.ChannelSMS] = provider.NewTwilioProvider(config.Twilio)
	}

	// Initialize WhatsApp provider
	if config.Whatsapp.WhatsappAccessToken != "" {
		logger.GetLogger().Infof("Initializing WhatsApp provider")
		factory.providers[constants.ChannelWhatsApp] = provider.NewWhatsAppProvider(config.Whatsapp)
	}

	// Initialize Zalo provider
	if config.Zalo.ZaloAppID != "" {
		zaloProvider, err := provider.NewZaloProvider(context.Background(), config.Zalo, zaloTokenRepo)
		if err != nil {
			logger.GetLogger().Errorf("Failed to create Zalo provider: %v", err)
		} else {
			factory.providers[constants.ChannelZalo] = zaloProvider
		}
	}

	// Always add webhook as fallback
	factory.providers[constants.DefaultSMSChannel] = provider.NewWebhookProvider()

	return factory, nil
}

// GetProvider returns the appropriate provider for the given channel
func (f *SMSProviderFactory) GetProvider(channel string) (provider.SMSProvider, error) {
	if provider, exists := f.providers[channel]; exists {
		return provider, nil
	}

	return nil, fmt.Errorf("provider for channel %s not found", channel)
}

// GetSupportedChannels returns all available channels
func (f *SMSProviderFactory) GetSupportedChannels() []string {
	channels := make([]string, 0, len(f.providers))
	for channel := range f.providers {
		channels = append(channels, channel)
	}
	return channels
}

// SMSService is the main service that orchestrates SMS sending
type SMSService struct {
	factory *SMSProviderFactory
}

// NewSMSService creates a new SMS service with the factory
func NewSMSService(config *conf.SmsConfiguration, zaloTokenRepo domainrepo.ZaloTokenRepository) (*SMSService, error) {
	factory, _ := NewSMSProviderFactory(config, zaloTokenRepo)
	return &SMSService{
		factory: factory,
	}, nil
}

// SendOTP sends an OTP through the specified channel
func (s *SMSService) SendOTP(ctx context.Context, tenantName, receiver, channel, otp string, ttl time.Duration) error {
	logger.GetLogger().Infof("Sending OTP to %s via channel %s", receiver, channel)

	provider, err := s.factory.GetProvider(channel)
	if err != nil {
		return fmt.Errorf("failed to get provider for channel %s: %w", channel, err)
	}
	return provider.SendOTP(ctx, tenantName, receiver, otp, ttl)
}

// GetSupportedChannels returns all supported channels
func (s *SMSService) GetSupportedChannels() []string {
	return s.factory.GetSupportedChannels()
}

func (s *SMSService) GetProvider(channel string) (provider.SMSProvider, error) {
	return s.factory.GetProvider(channel)
}
