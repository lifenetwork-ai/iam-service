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
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms/common"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

const (
	DefaultTokenDuration = time.Hour
	MaxRetryTime         = 30 * time.Second
	MaxRetries           = 3
	TokenInvalidError    = -124
	SuccessError         = 0
)

// ZaloProvider handles messages through Zalo
type ZaloProvider struct {
	client                 *client.ZaloClient
	config                 conf.ZaloConfiguration
	tokenRepo              domainrepo.ZaloTokenRepository
	zaloTokenCryptoService *common.ZaloTokenCrypto
}

// Core public methods

func (z *ZaloProvider) SendOTP(ctx context.Context, tenantName, receiver, otp string, ttl time.Duration) error {
	logger.GetLogger().Infof("Sending OTP to %s via Zalo for tenant %s", receiver, tenantName)

	operation := func() error {
		return z.attemptSendOTP(ctx, receiver, otp)
	}

	if err := z.retryWithBackoff(operation); err != nil {
		return fmt.Errorf("failed to send OTP after retries: %w", err)
	}

	logger.GetLogger().Info("Zalo OTP sent successfully")
	return nil
}

// Private helper methods

func (z *ZaloProvider) initializeToken(ctx context.Context) error {
	token, err := z.getOrCreateToken(ctx)
	if err != nil {
		return err
	}

	if z.isTokenExpired(token) {
		return z.refreshAndSaveToken(ctx)
	}

	return nil
}

func (z *ZaloProvider) getOrCreateToken(ctx context.Context) (*domain.ZaloToken, error) {
	token, err := z.tokenRepo.Get(ctx)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get Zalo token from DB: %w", err)
	}

	if token != nil {
		return token, nil
	}

	return z.createInitialToken(ctx)
}

func (z *ZaloProvider) createInitialToken(ctx context.Context) (*domain.ZaloToken, error) {
	if z.config.ZaloAccessToken == "" || z.config.ZaloRefreshToken == "" {
		return nil, fmt.Errorf("no Zalo tokens found in DB or config")
	}

	token := &domain.ZaloToken{
		AccessToken:  z.config.ZaloAccessToken,
		RefreshToken: z.config.ZaloRefreshToken,
		UpdatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(DefaultTokenDuration),
	}

	if err := z.tokenRepo.Save(ctx, token); err != nil {
		return nil, fmt.Errorf("failed to persist initial Zalo token to DB: %w", err)
	}

	return token, nil
}

func (z *ZaloProvider) createClient(ctx context.Context) error {
	token, err := z.tokenRepo.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get token for client creation: %w", err)
	}

	z.client, err = client.NewZaloClient(
		ctx,
		z.config.ZaloBaseURL,
		z.config.ZaloSecretKey,
		z.config.ZaloAppID,
		token.AccessToken,
		token.RefreshToken,
	)
	if err != nil {
		return fmt.Errorf("failed to create Zalo client: %w", err)
	}

	return nil
}

func (z *ZaloProvider) isTokenExpired(token *domain.ZaloToken) bool {
	return token.ExpiresAt.Before(time.Now())
}

func (z *ZaloProvider) attemptSendOTP(ctx context.Context, receiver, otp string) error {
	resp, err := z.client.SendOTP(ctx, receiver, otp, z.config.ZaloTemplateID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to send OTP via Zalo: %v", err)
		return fmt.Errorf("failed to send OTP via Zalo: %w", err)
	}

	return z.handleAPIResponse(ctx, resp)
}

func (z *ZaloProvider) handleAPIResponse(ctx context.Context, resp *client.ZaloTemplateResponse) error {
	switch resp.Error {
	case SuccessError:
		return nil
	case TokenInvalidError:
		logger.GetLogger().Info("Access token invalid, refreshing token")
		if err := z.refreshAndSaveToken(ctx); err != nil {
			return err
		}
		return fmt.Errorf("access token was invalid and refreshed, retrying")
	default:
		return backoff.Permanent(fmt.Errorf("zalo api error: code %d, message: %s", resp.Error, resp.Message))
	}
}

func (z *ZaloProvider) RefreshToken(ctx context.Context) error {
	logger.GetLogger().Info("Refreshing Zalo token")

	operation := func() error {
		return z.refreshAndSaveToken(ctx)
	}

	if err := z.retryWithBackoff(operation); err != nil {
		return fmt.Errorf("failed to refresh Zalo token: %w", err)
	}

	logger.GetLogger().Info("Zalo token refreshed successfully")
	return nil
}

func (z *ZaloProvider) GetChannelType() string {
	return constants.ChannelZalo
}

func (z *ZaloProvider) HealthCheck(ctx context.Context) error {
	// Implement Zalo health check if available
	return nil
}

// Constructor and initialization

func NewZaloProvider(ctx context.Context, config conf.ZaloConfiguration, tokenRepo domainrepo.ZaloTokenRepository) (*ZaloProvider, error) {
	if tokenRepo == nil {
		return nil, fmt.Errorf("tokenRepo is nil")
	}

	zaloTokenCryptoService := common.NewZaloTokenCrypto()

	provider := &ZaloProvider{
		config:                 config,
		tokenRepo:              tokenRepo,
		zaloTokenCryptoService: zaloTokenCryptoService,
	}

	if err := provider.initializeToken(ctx); err != nil {
		return nil, err
	}

	if err := provider.createClient(ctx); err != nil {
		return nil, err
	}

	return provider, nil
}

// Private helper methods

func (z *ZaloProvider) refreshAndSaveToken(ctx context.Context) error {
	resp, err := z.client.RefreshAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh access token: %w", err)
	}

	if err := z.client.UpdateTokens(ctx, resp); err != nil {
		return fmt.Errorf("failed to update Zalo client tokens: %w", err)
	}

	return z.saveTokenToDB(ctx, resp)
}

func (z *ZaloProvider) saveTokenToDB(ctx context.Context, resp *client.ZaloTokenRefreshResponse) error {
	expiresIn, err := strconv.Atoi(resp.ExpiresIn)
	if err != nil {
		return fmt.Errorf("failed to convert expiresIn to int: %w", err)
	}

	dbToken := &domain.ZaloToken{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(expiresIn) * time.Second),
		UpdatedAt:    time.Now(),
	}

	encryptedToken, err := z.zaloTokenCryptoService.Encrypt(ctx, dbToken)
	if err != nil {
		return fmt.Errorf("failed to encrypt tokens: %w", err)
	}

	if err := z.tokenRepo.Save(ctx, encryptedToken); err != nil {
		return fmt.Errorf("failed to save refreshed tokens to DB: %w", err)
	}

	return nil
}

func (z *ZaloProvider) retryWithBackoff(operation func() error) error {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = MaxRetryTime

	return backoff.Retry(operation, backoff.WithMaxRetries(b, MaxRetries))
}
