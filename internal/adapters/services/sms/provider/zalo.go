package provider

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

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
	AssumedTokenLifetime = time.Hour
	MaxRetryTime         = 30 * time.Second
	MaxRetries           = 3
	TokenInvalidError    = -124
	SuccessCode          = 0
	// TokenRefreshGrace defines how soon before expiry we proactively refresh
	TokenRefreshGrace = 5 * time.Minute
)

// ZaloProvider handles messages through Zalo
type ZaloProvider struct {
	client                 *client.ZaloClient
	config                 conf.ZaloConfiguration
	tokenRepo              domainrepo.ZaloTokenRepository
	zaloTokenCryptoService *common.ZaloTokenCrypto
	mu                     sync.Mutex
}

func NewZaloProvider(ctx context.Context, config conf.ZaloConfiguration, tokenRepo domainrepo.ZaloTokenRepository) (SMSProvider, error) {
	// Validate inputs
	if tokenRepo == nil {
		return nil, fmt.Errorf("tokenRepo is required")
	}

	if err := validateZaloConfig(config); err != nil {
		return nil, fmt.Errorf("invalid Zalo configuration: %w", err)
	}

	zaloTokenCryptoService := common.NewZaloTokenCrypto(conf.GetConfiguration().DbEncryptionKey)

	provider := &ZaloProvider{
		config:                 config,
		tokenRepo:              tokenRepo,
		zaloTokenCryptoService: zaloTokenCryptoService,
	}

	if err := provider.createClient(ctx); err != nil {
		return nil, fmt.Errorf("failed to create Zalo client: %w", err)
	}

	return provider, nil
}

func validateZaloConfig(config conf.ZaloConfiguration) error {
	if config.ZaloBaseURL == "" {
		return fmt.Errorf("ZaloBaseURL is required")
	}
	if config.ZaloSecretKey == "" {
		return fmt.Errorf("ZaloSecretKey is required")
	}
	if config.ZaloAppID == "" {
		return fmt.Errorf("ZaloAppID is required")
	}
	// Add template ID validation if required
	return nil
}

// Core public methods
func (z *ZaloProvider) SendOTP(ctx context.Context, tenantName, receiver, otp string, ttl time.Duration) error {
	logger.GetLogger().Infof("Sending OTP to %s via Zalo for tenant %s", receiver, tenantName)

	// Sync latest tokens from storage and proactively refresh if near expiry
	if err := z.syncTokensBeforeSend(ctx); err != nil {
		return fmt.Errorf("failed to prepare tokens before send: %w", err)
	}

	operation := func() error {
		return z.attemptSendOTP(ctx, receiver, otp)
	}

	if err := z.retryWithBackoff(operation); err != nil {
		return fmt.Errorf("failed to send OTP after retries: %w", err)
	}

	logger.GetLogger().Info("Zalo OTP sent successfully")
	return nil
}

// HealthCheck checks if the Zalo client is healthy by getting all templates using the client
func (z *ZaloProvider) HealthCheck(ctx context.Context) error {
	resp, err := z.client.GetAllTemplates(ctx)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get all templates: %v", err)
		return fmt.Errorf("failed to get all templates: %v", err)
	}
	// Parse the error - it can be int, float64, or string
	var errCode int
	switch v := resp.Error.(type) {
	case int:
		errCode = v
	case float64:
		errCode = int(v)
	case string:
		// If it's a string, we assume it's an error (non-zero)
		if v == "0" || v == "" {
			errCode = SuccessCode
		} else {
			errCode = -1 // Non-zero error code
		}
	default:
		logger.GetLogger().Errorf("Unexpected error type from response: %T, value: %v", resp.Error, resp.Error)
		return fmt.Errorf("unexpected error type from response: %T, value: %v", resp.Error, resp.Error)
	}
	if errCode != SuccessCode {
		logger.GetLogger().Errorf("Failed to get all templates: %v", resp.Message)
		return fmt.Errorf("failed to get all templates: %v", resp.Message)
	}

	return nil
}

// RefreshToken refreshes the Zalo token
// If refreshToken is not provided, it will use the refresh token from the client
func (z *ZaloProvider) RefreshToken(ctx context.Context, refreshToken string) error {
	logger.GetLogger().Info("Refreshing Zalo token")
	if refreshToken == "" {
		refreshToken = z.client.GetRefreshToken(ctx)
	}
	operation := func() error {
		z.mu.Lock()
		defer z.mu.Unlock()
		return z.refreshAndSaveToken(ctx, refreshToken)
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

// Private helper methods

// getOrCreateToken gets the Zalo token from the repository or creates a new one if it doesn't exist
func (z *ZaloProvider) getOrCreateToken(ctx context.Context) (*domain.ZaloToken, error) {
	token, _ := z.getTokenFromDB(ctx)

	if token != nil {
		return token, nil
	}

	// If no token is found, use the default tokens from the config to seed the database
	return z.createInitialToken(ctx)
}

// createInitialToken creates a new Zalo token, used when no token is found in the repository
func (z *ZaloProvider) createInitialToken(ctx context.Context) (*domain.ZaloToken, error) {
	if z.config.ZaloAccessToken == "" || z.config.ZaloRefreshToken == "" {
		return nil, fmt.Errorf("no Zalo tokens found in config")
	}

	token := &domain.ZaloToken{
		ID:           1,
		AccessToken:  z.config.ZaloAccessToken,
		RefreshToken: z.config.ZaloRefreshToken,
		UpdatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(AssumedTokenLifetime),
	}

	encryptedToken, err := z.zaloTokenCryptoService.Encrypt(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt initial Zalo token: %w", err)
	}

	if err := z.tokenRepo.Save(ctx, encryptedToken); err != nil {
		return nil, fmt.Errorf("failed to persist initial Zalo token to DB: %w", err)
	}

	return token, nil
}

func (z *ZaloProvider) attemptSendOTP(ctx context.Context, receiver, otp string) error {
	resp, err := z.client.SendOTP(ctx, receiver, otp, z.config.ZaloTemplateID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to send OTP via Zalo: %v", err)
		return fmt.Errorf("failed to send OTP via Zalo: %w", err)
	}

	return z.handleAPIResponse(ctx, resp)
}

func (z *ZaloProvider) handleAPIResponse(ctx context.Context, resp *client.ZaloSendNotificationResponse) error {
	switch resp.Error {
	case SuccessCode:
		return nil
	case TokenInvalidError:
		logger.GetLogger().Info("Refresh token invalid, trying to refresh token")
		z.mu.Lock()
		defer z.mu.Unlock()
		if err := z.refreshAndSaveToken(ctx, z.config.ZaloRefreshToken); err != nil {
			return err
		}
		return fmt.Errorf("access token was invalid and refreshed, trying to send OTP again")
	default:
		return backoff.Permanent(fmt.Errorf("zalo api error: code %d, message: %s", resp.Error, resp.Message))
	}
}

// refreshAndSaveToken refreshes tokens via the client and persists them.
// Caller MUST hold z.mu.
func (z *ZaloProvider) refreshAndSaveToken(ctx context.Context, refreshToken string) error {
	// Ensure client exists before attempting refresh
	if z.client == nil {
		cli, err := client.NewZaloClient(
			ctx,
			z.config.ZaloBaseURL,
			z.config.ZaloSecretKey,
			z.config.ZaloAppID,
			"", // no access token needed for refresh
			refreshToken,
		)
		if err != nil {
			return fmt.Errorf("failed to bootstrap Zalo client for refresh: %w", err)
		}
		z.client = cli
	}
	resp, err := z.client.RefreshAccessToken(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("failed to refresh access token: %w", err)
	}

	if err := z.client.UpdateTokens(ctx, resp); err != nil {
		return fmt.Errorf("failed to update Zalo client tokens: %w", err)
	}

	return z.saveTokenToDB(ctx, resp)
}

// getTokenFromDB gets the Zalo token from the database and decrypts it
func (z *ZaloProvider) getTokenFromDB(ctx context.Context) (*domain.ZaloToken, error) {
	token, err := z.tokenRepo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Zalo token from DB: %w", err)
	}
	if token == nil {
		return nil, fmt.Errorf("no Zalo token found in DB")
	}
	// decrypt the token
	decryptedToken, err := z.zaloTokenCryptoService.Decrypt(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt Zalo token: %w", err)
	}
	return decryptedToken, nil
}

// saveTokenToDB encrypts the newly refreshed Zalo token and saves it to the database
func (z *ZaloProvider) saveTokenToDB(ctx context.Context, resp *client.ZaloTokenRefreshResponse) error {
	expiresIn, err := strconv.Atoi(resp.ExpiresIn)
	if err != nil {
		return fmt.Errorf("failed to convert expiresIn to int: %w", err)
	}

	dbToken := &domain.ZaloToken{
		ID:           1,
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
	b.InitialInterval = 2 * time.Second

	return backoff.Retry(operation, backoff.WithMaxRetries(b, MaxRetries))
}

// syncTokensBeforeSend ensures the client has the latest tokens from storage
// and proactively refreshes if the token is near expiry.
func (z *ZaloProvider) syncTokensBeforeSend(ctx context.Context) error {
	z.mu.Lock()
	defer z.mu.Unlock()
	token, err := z.getTokenFromDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get token for sync: %w", err)
	}

	if token == nil {
		// Nothing in storage yet; rely on current client tokens
		return nil
	}

	// If token is near expiry, refresh and persist new tokens
	if time.Until(token.ExpiresAt) <= TokenRefreshGrace {
		return z.refreshAndSaveToken(ctx, token.RefreshToken)
	}

	// Update the client to use the latest tokens from storage.
	_ = z.client.UpdateTokens(ctx, &client.ZaloTokenRefreshResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.ExpiresAt.Format(time.RFC3339),
	})

	return nil
}

// isTokenExpired checks if the Zalo token is expired
func (z *ZaloProvider) isTokenExpired(token *domain.ZaloToken) bool {
	return token.ExpiresAt.Before(time.Now())
}

// createClient creates a new Zalo client
func (z *ZaloProvider) createClient(ctx context.Context) error {
	token, err := z.getOrCreateToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get token for client creation: %w", err)
	}

	// If token is expired, refresh it first
	if z.isTokenExpired(token) {
		if err := z.refreshAndSaveToken(ctx, token.RefreshToken); err != nil {
			return fmt.Errorf("failed to refresh expired token: %w", err)
		}
		// Get the refreshed token
		token, err = z.getOrCreateToken(ctx)
		if err != nil {
			return fmt.Errorf("failed to get refreshed token: %w", err)
		}
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
