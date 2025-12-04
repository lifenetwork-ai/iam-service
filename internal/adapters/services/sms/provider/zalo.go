package provider

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/google/uuid"
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
// In multi-tenant mode, clients are created on-demand per request
type ZaloProvider struct {
	config                 conf.ZaloConfiguration
	tokenRepo              domainrepo.ZaloTokenRepository
	tenantRepo             domainrepo.TenantRepository
	zaloTokenCryptoService *common.ZaloTokenCrypto
	mu                     sync.Mutex
}

func NewZaloProvider(ctx context.Context, config conf.ZaloConfiguration, tokenRepo domainrepo.ZaloTokenRepository, tenantRepo domainrepo.TenantRepository) (SMSProvider, error) {
	// Validate inputs
	if tokenRepo == nil {
		return nil, fmt.Errorf("tokenRepo is required")
	}
	if tenantRepo == nil {
		return nil, fmt.Errorf("tenantRepo is required")
	}

	if err := validateZaloConfig(config); err != nil {
		return nil, fmt.Errorf("invalid Zalo configuration: %w", err)
	}

	zaloTokenCryptoService := common.NewZaloTokenCrypto(conf.GetConfiguration().DbEncryptionKey)

	provider := &ZaloProvider{
		config:                 config,
		tokenRepo:              tokenRepo,
		tenantRepo:             tenantRepo,
		zaloTokenCryptoService: zaloTokenCryptoService,
	}

	// Don't create client at initialization - will be created per-tenant on first use
	return provider, nil
}

func NewZaloProviderWithRefresh(
	ctx context.Context,
	config conf.ZaloConfiguration,
	tokenRepo domainrepo.ZaloTokenRepository,
	tenantRepo domainrepo.TenantRepository,
	refreshToken string,
) (*ZaloProvider, error) {
	// This constructor is deprecated in multi-tenant mode
	// Each tenant has their own tokens configured via the admin API
	return nil, fmt.Errorf("NewZaloProviderWithRefresh is deprecated in multi-tenant mode")
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

// getTenantIDFromName converts tenant name to tenant ID
func (z *ZaloProvider) getTenantIDFromName(_ context.Context, tenantName string) (uuid.UUID, error) {
	tenant, err := z.tenantRepo.GetByName(tenantName)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get tenant by name %s: %w", tenantName, err)
	}
	if tenant == nil {
		return uuid.Nil, fmt.Errorf("tenant not found: %s", tenantName)
	}
	return tenant.ID, nil
}

// Core public methods
func (z *ZaloProvider) SendOTP(ctx context.Context, tenantName, receiver, otp string, ttl time.Duration) error {
	logger.GetLogger().Infof("Sending OTP to %s via Zalo for tenant %s", receiver, tenantName)

	// Convert tenant name to ID
	tenantID, err := z.getTenantIDFromName(ctx, tenantName)
	if err != nil {
		return fmt.Errorf("failed to resolve tenant: %w", err)
	}

	// Get token for this tenant
	token, err := z.getTokenFromDB(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get token for tenant: %w", err)
	}

	// Proactively refresh if near expiry
	if time.Until(token.ExpiresAt) <= TokenRefreshGrace {
		logger.GetLogger().Infof("Token near expiry, proactively refreshing for tenant %s", tenantName)
		if _, err := z.refreshAndSaveToken(ctx, token.RefreshToken, tenantID); err != nil {
			logger.GetLogger().Warnf("Failed to proactively refresh token: %v", err)
			// Continue anyway, might still work
		} else {
			// Re-fetch the refreshed token
			token, err = z.getTokenFromDB(ctx, tenantID)
			if err != nil {
				return fmt.Errorf("failed to get refreshed token: %w", err)
			}
		}
	}

	// Create client for this tenant
	cli, err := client.NewZaloClient(
		ctx,
		z.config.ZaloBaseURL,
		token.SecretKey,
		token.AppID,
		token.AccessToken,
		token.RefreshToken,
	)
	if err != nil {
		return fmt.Errorf("failed to create Zalo client: %w", err)
	}

	operation := func() error {
		return z.attemptSendOTP(ctx, cli, receiver, otp, tenantID)
	}

	if err := z.retryWithBackoff(operation); err != nil {
		return fmt.Errorf("failed to send OTP after retries: %w", err)
	}

	logger.GetLogger().Info("Zalo OTP sent successfully")
	return nil
}

// HealthCheck is deprecated in multi-tenant mode
// Use the sms_token use case ZaloHealthCheck method instead
func (z *ZaloProvider) HealthCheck(ctx context.Context) error {
	return fmt.Errorf("HealthCheck is deprecated in multi-tenant mode, use SMS token use case instead")
}

// RefreshToken refreshes the Zalo token for a specific tenant
func (z *ZaloProvider) RefreshToken(ctx context.Context, refreshToken string) error {
	// This method is deprecated in multi-tenant mode
	// Token refresh should be done via the use case layer with tenant context
	return fmt.Errorf("RefreshToken is deprecated in multi-tenant mode, use SMS token use case instead")
}

func (z *ZaloProvider) GetChannelType() string {
	return constants.ChannelZalo
}

func (z *ZaloProvider) attemptSendOTP(ctx context.Context, cli *client.ZaloClient, receiver, otp string, tenantID uuid.UUID) error {
	resp, err := cli.SendOTP(ctx, receiver, otp, z.config.ZaloTemplateID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to send OTP via Zalo: %v", err)
		return fmt.Errorf("failed to send OTP via Zalo: %w", err)
	}

	return z.handleAPIResponse(ctx, cli, resp, tenantID)
}

func (z *ZaloProvider) handleAPIResponse(ctx context.Context, cli *client.ZaloClient, resp *client.ZaloSendNotificationResponse, tenantID uuid.UUID) error {
	switch resp.Error {
	case SuccessCode:
		return nil
	case TokenInvalidError:
		logger.GetLogger().Info("Access token invalid, trying to refresh token")
		z.mu.Lock()
		defer z.mu.Unlock()

		// Get current refresh token
		refreshToken := cli.GetRefreshToken(ctx)
		// Refresh tokens, persist to DB, and update in-memory client so immediate retry uses fresh tokens
		newTokens, err := z.refreshAndSaveToken(ctx, refreshToken, tenantID)
		if err != nil {
			return err
		}
		if err := cli.UpdateTokens(ctx, newTokens); err != nil {
			return fmt.Errorf("failed to update in-memory client tokens: %w", err)
		}
		return fmt.Errorf("access token was invalid and refreshed, retry needed")
	default:
		return backoff.Permanent(fmt.Errorf("zalo api error: code %d, message: %s", resp.Error, resp.Message))
	}
}

// refreshAndSaveToken refreshes tokens via the client and persists them.
// Caller MUST hold z.mu.
func (z *ZaloProvider) refreshAndSaveToken(ctx context.Context, refreshToken string, tenantID uuid.UUID) (*client.ZaloTokenRefreshResponse, error) {
	// Get token details to get AppID and SecretKey
	token, err := z.getTokenFromDB(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token details: %w", err)
	}

	// Create a temporary client for refresh
	cli, err := client.NewZaloClient(
		ctx,
		z.config.ZaloBaseURL,
		token.SecretKey,
		token.AppID,
		"", // no access token needed for refresh
		refreshToken,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for refresh: %w", err)
	}

	resp, err := cli.RefreshAccessToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh access token: %w", err)
	}

	if err := z.saveTokenToDB(ctx, resp, tenantID); err != nil {
		return nil, err
	}
	return resp, nil
}

// getTokenFromDB gets the Zalo token from the database and decrypts it
func (z *ZaloProvider) getTokenFromDB(ctx context.Context, tenantID uuid.UUID) (*domain.ZaloToken, error) {
	token, err := z.tokenRepo.Get(ctx, tenantID)
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
func (z *ZaloProvider) saveTokenToDB(ctx context.Context, resp *client.ZaloTokenRefreshResponse, tenantID uuid.UUID) error {
	expiresIn, err := strconv.Atoi(resp.ExpiresIn)
	if err != nil {
		return fmt.Errorf("failed to convert expiresIn to int: %w", err)
	}

	// First get and decrypt existing token to preserve ID and real AppID/SecretKey
	existingEncrypted, err := z.tokenRepo.Get(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get existing token: %w", err)
	}
	existingDecrypted, err := z.zaloTokenCryptoService.Decrypt(ctx, existingEncrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt existing token: %w", err)
	}

	dbToken := &domain.ZaloToken{
		ID:           existingEncrypted.ID,
		TenantID:     tenantID,
		AppID:        existingDecrypted.AppID,
		SecretKey:    existingDecrypted.SecretKey,
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
