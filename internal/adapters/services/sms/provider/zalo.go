package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/constants"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms/client"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

// ZaloProvider handles messages through Zalo
type ZaloProvider struct {
	client    *client.ZaloClient
	config    conf.ZaloConfiguration
	tokenRepo domainrepo.ZaloTokenRepository
}

func NewZaloProvider(ctx context.Context, config conf.ZaloConfiguration, tokenRepo domainrepo.ZaloTokenRepository) (*ZaloProvider, error) {
	client, err := client.NewZaloClient(
		ctx,
		config.ZaloBaseURL,
		config.ZaloSecretKey,
		config.ZaloAppID,
		config.ZaloAccessToken,
		config.ZaloRefreshToken,
		tokenRepo,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Zalo client: %w", err)
	}

	return &ZaloProvider{
		client:    client,
		config:    config,
		tokenRepo: tokenRepo,
	}, nil
}

func (z *ZaloProvider) SendOTP(ctx context.Context, tenantName, receiver, otp string, ttl time.Duration) error {
	logger.GetLogger().Infof("Sending OTP to %s via Zalo for tenant %s", receiver, tenantName)

	resp, err := z.client.SendOTP(ctx, receiver, otp, z.config.ZaloTemplateID)
	if err != nil {
		return fmt.Errorf("failed to send OTP via Zalo: %w", err)
	}

	logger.GetLogger().Infof("Zalo OTP sent successfully: %+v", resp)
	return nil
}

func (z *ZaloProvider) RefreshToken(ctx context.Context) error {
	logger.GetLogger().Infof("Refreshing Zalo token")
	var resp *client.ZaloTokenRefreshResponse
	// Use the client's refresh token functionality
	backoff.Retry(func() error {
		var err error
		resp, err = z.client.RefreshAccessToken(ctx)
		if err != nil {
			return fmt.Errorf("failed to refresh Zalo token: %w", err)
		}
		return nil
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 3))

	// Update the client's tokens
	if err := z.client.UpdateTokens(ctx, resp); err != nil {
		return fmt.Errorf("failed to update Zalo client tokens: %w", err)
	}

	// Persist new tokens in DB
	dbToken := &domain.ZaloToken{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		UpdatedAt:    time.Now(),
	}
	if err := z.tokenRepo.Save(ctx, dbToken); err != nil {
		return fmt.Errorf("failed to save refreshed tokens to DB: %w", err)
	}

	logger.GetLogger().Infof("Zalo token refreshed successfully: %+v", resp)
	return nil
}

func (z *ZaloProvider) GetChannelType() string {
	return constants.ChannelZalo
}

func (z *ZaloProvider) HealthCheck(ctx context.Context) error {
	// Implement Zalo health check if available
	return nil
}
