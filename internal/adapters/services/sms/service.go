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

// SMSProvider defines the interface that all SMS providers must implement
type SMSProvider interface {
	SendOTP(ctx context.Context, tenantName, receiver, otp string, ttl time.Duration) error
	RefreshToken(ctx context.Context) error
	GetChannelType() string
	HealthCheck(ctx context.Context) error
}

// SMSProviderFactory manages and creates SMS providers
type SMSProviderFactory struct {
	providers map[string]SMSProvider
}

// NewSMSProviderFactory creates a new factory with all configured providers
func NewSMSProviderFactory(config *conf.SmsConfiguration, zaloTokenRepo domainrepo.ZaloTokenRepository) (*SMSProviderFactory, error) {
	factory := &SMSProviderFactory{
		providers: make(map[string]SMSProvider),
	}

	// Initialize Twilio provider
	if config.Twilio.TwilioAccountSID != "" {
		factory.providers[constants.ChannelSMS] = provider.NewTwilioProvider(config.Twilio)
	}

	// Initialize WhatsApp provider
	if config.Whatsapp.WhatsappAccessToken != "" {
		factory.providers[constants.ChannelWhatsApp] = provider.NewWhatsAppProvider(config.Whatsapp)
	}

	// Initialize Zalo provider
	if config.Zalo.ZaloAppID != "" {
		zaloProvider, err := provider.NewZaloProvider(context.Background(), config.Zalo, zaloTokenRepo)
		if err != nil {
			logger.GetLogger().Errorf("Failed to create Zalo provider: %v", err)
			// Don't return error, just log and continue without Zalo
		} else {
			factory.providers[constants.ChannelZalo] = zaloProvider
		}
	}

	// Always add webhook as fallback
	factory.providers["webhook"] = provider.NewWebhookProvider()

	return factory, nil
}

// GetProvider returns the appropriate provider for the given channel
func (f *SMSProviderFactory) GetProvider(channel string) (SMSProvider, error) {
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

// RefreshAllTokens refreshes tokens for all providers that support it
func (f *SMSProviderFactory) RefreshAllTokens(ctx context.Context) error {
	var errors []error
	for channel, provider := range f.providers {
		if err := provider.RefreshToken(ctx); err != nil {
			logger.GetLogger().Errorf("Failed to refresh token for %s: %v", channel, err)
			errors = append(errors, fmt.Errorf("channel %s: %w", channel, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to refresh tokens: %v", errors)
	}
	return nil
}

// HealthCheckAll performs health checks on all providers
func (f *SMSProviderFactory) HealthCheckAll(ctx context.Context) map[string]error {
	results := make(map[string]error)
	for channel, provider := range f.providers {
		results[channel] = provider.HealthCheck(ctx)
	}
	return results
}

// SMSService is the main service that orchestrates SMS sending
type SMSService struct {
	factory *SMSProviderFactory
}

// NewSMSService creates a new SMS service with the factory
func NewSMSService(config *conf.SmsConfiguration, zaloTokenRepo domainrepo.ZaloTokenRepository) (*SMSService, error) {
	factory, err := NewSMSProviderFactory(config, zaloTokenRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to create SMS provider factory: %w", err)
	}

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

// RefreshTokens refreshes tokens for all providers
func (s *SMSService) RefreshTokens(ctx context.Context) error {
	return s.factory.RefreshAllTokens(ctx)
}

// HealthCheck performs health checks on all providers
func (s *SMSService) HealthCheck(ctx context.Context) map[string]error {
	return s.factory.HealthCheckAll(ctx)
}

func (s *SMSService) GetProvider(channel string) (SMSProvider, error) {
	return s.factory.GetProvider(channel)
}
