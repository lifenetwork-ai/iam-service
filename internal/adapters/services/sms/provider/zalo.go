package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

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
	if tokenRepo == nil {
		return nil, fmt.Errorf("tokenRepo is nil")
	}

	// Try to get token from DB
	token, err := tokenRepo.Get(ctx)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get Zalo token from DB: %w", err)
	}

	// Fallback to config if DB token not found
	if token == nil {
		if config.ZaloAccessToken == "" || config.ZaloRefreshToken == "" {
			return nil, fmt.Errorf("no Zalo tokens found in DB or config")
		}

		// assume that the token provided via env is fresh and valid for at least 1 hour
		// 1h is sufficient for the first time setup, the next interaction will refresh the token
		token = &domain.ZaloToken{
			AccessToken:  config.ZaloAccessToken,
			RefreshToken: config.ZaloRefreshToken,
			UpdatedAt:    time.Now(),
			ExpiresAt:    time.Now().Add(time.Hour),
		}

		// Persist initial token to DB if DB is empty
		if err := tokenRepo.Save(ctx, token); err != nil {
			return nil, fmt.Errorf("failed to persist initial Zalo token to DB: %w", err)
		}
	}

	client, err := client.NewZaloClient(
		ctx,
		config.ZaloBaseURL,
		config.ZaloSecretKey,
		config.ZaloAppID,
		token.AccessToken,
		token.RefreshToken,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Zalo client: %w", err)
	}

	// if token is expired, refresh it
	if token.ExpiresAt.Before(time.Now()) {
		resp, err := client.RefreshAccessToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh Zalo token: %w", err)
		}
		if err := client.UpdateTokens(ctx, resp); err != nil {
			return nil, fmt.Errorf("failed to update Zalo client tokens: %w", err)
		}

		expiresIn, err := strconv.Atoi(resp.ExpiresIn)
		if err != nil {
			return nil, fmt.Errorf("failed to convert expiresIn to int: %w", err)
		}
		if err := tokenRepo.Save(ctx, &domain.ZaloToken{
			AccessToken:  resp.AccessToken,
			RefreshToken: resp.RefreshToken,
			ExpiresAt:    time.Now().Add(time.Duration(expiresIn) * time.Second),
			UpdatedAt:    time.Now(),
		}); err != nil {
			return nil, fmt.Errorf("failed to save refreshed tokens to DB: %w", err)
		}
	}

	return &ZaloProvider{
		client:    client,
		config:    config,
		tokenRepo: tokenRepo,
	}, nil
}

func (z *ZaloProvider) SendOTP(ctx context.Context, tenantName, receiver, otp string, ttl time.Duration) error {
	logger.GetLogger().Infof("Sending OTP to %s via Zalo for tenant %s", receiver, tenantName)

	var resp *client.ZaloTemplateResponse

	// Use exponential backoff for sending OTP with token refresh capability
	operation := func() error {
		var err error
		resp, err = z.client.SendOTP(ctx, receiver, otp, z.config.ZaloTemplateID)
		if err != nil {
			return fmt.Errorf("failed to send OTP via Zalo: %w", err)
		}

		// Check for API errors
		// {
		// 	"error": -124,
		// 	"message": "Access token invalid"
		// }
		// If access token is invalid, refresh it and try again
		if resp.Error == -124 {
			logger.GetLogger().Infof("Access token invalid, refreshing token")

			refreshTokenResp, err := z.client.RefreshAccessToken(ctx)
			if err != nil {
				return fmt.Errorf("failed to refresh access token: %w", err)
			}

			// Update client tokens
			if err := z.client.UpdateTokens(ctx, refreshTokenResp); err != nil {
				return fmt.Errorf("failed to update Zalo client tokens: %w", err)
			}

			// Parse and save token expiration to database
			expiresIn, err := strconv.Atoi(refreshTokenResp.ExpiresIn)
			if err != nil {
				return fmt.Errorf("failed to convert expiresIn to int: %w", err)
			}

			dbToken := &domain.ZaloToken{
				AccessToken:  refreshTokenResp.AccessToken,
				RefreshToken: refreshTokenResp.RefreshToken,
				ExpiresAt:    time.Now().Add(time.Duration(expiresIn) * time.Second),
				UpdatedAt:    time.Now(),
			}
			if err := z.tokenRepo.Save(ctx, dbToken); err != nil {
				return fmt.Errorf("failed to save refreshed tokens to DB: %w", err)
			}

			// Return a retryable error to trigger backoff retry
			return fmt.Errorf("access token was invalid and refreshed, retrying")
		} else if resp.Error != 0 {
			// Handle other API errors as permanent errors (no retry)
			return backoff.Permanent(fmt.Errorf("Zalo API error: code %d, message: %s", resp.Error, resp.Message))
		}

		return nil
	}

	// Configure exponential backoff with max retries
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 30 * time.Second // Maximum total time for retries

	err := backoff.Retry(operation, backoff.WithMaxRetries(b, 3))
	if err != nil {
		return fmt.Errorf("failed to send OTP after retries: %w", err)
	}

	logger.GetLogger().Infof("Zalo OTP sent successfully: %+v", resp)
	return nil
}

func (z *ZaloProvider) RefreshToken(ctx context.Context) error {
	logger.GetLogger().Infof("Refreshing Zalo token")
	var resp *client.ZaloTokenRefreshResponse
	// Use the client's refresh token functionality
	err := backoff.Retry(func() error {
		var err error
		resp, err = z.client.RefreshAccessToken(ctx)
		if err != nil {
			return fmt.Errorf("failed to refresh Zalo token: %w", err)
		}
		return nil
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 3))
	if err != nil {
		return fmt.Errorf("failed to refresh Zalo token: %w", err)
	}

	// Update the client's tokens
	if err := z.client.UpdateTokens(ctx, resp); err != nil {
		return fmt.Errorf("failed to update Zalo client tokens: %w", err)
	}

	// TODO: refactor this duplicated logic
	expiresIn, err := strconv.Atoi(resp.ExpiresIn)
	if err != nil {
		return fmt.Errorf("failed to convert expiresIn to int: %w", err)
	}

	// Persist new tokens in DB
	dbToken := &domain.ZaloToken{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		UpdatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Duration(expiresIn) * time.Second),
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
