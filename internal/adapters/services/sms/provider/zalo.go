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
	// TokenCacheTTL defines how long a decrypted token stays in cache
	TokenCacheTTL = 30 * time.Second
)

// ZaloProvider handles messages through Zalo in multi-tenant mode.
// Clients are constructed per request. A small cache keeps decrypted tokens to
// avoid repeated decrypt/DB hits. Cache is invalidated/updated on refresh.
type ZaloProvider struct {
	config                 conf.ZaloConfiguration
	tokenRepo              domainrepo.ZaloTokenRepository
	tenantRepo             domainrepo.TenantRepository
	zaloTokenCryptoService *common.ZaloTokenCrypto
	mu                     sync.Mutex
	// tokenCache keeps decrypted tokens with a short TTL
	tokenCache map[uuid.UUID]*tokenCacheEntry
}

type tokenCacheEntry struct {
	token    *domain.ZaloToken
	loadedAt time.Time
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
		tokenCache:             make(map[uuid.UUID]*tokenCacheEntry),
	}

	return provider, nil
}

func validateZaloConfig(config conf.ZaloConfiguration) error {
	if config.ZaloBaseURL == "" {
		return fmt.Errorf("ZaloBaseURL is required")
	}
	// Per-tenant credentials are stored in DB; no env-based AppID/SecretKey validation here
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

	// Get decrypted token (with small cache)
	token, err := z.getTokenCached(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get token for tenant: %w", err)
	}

	// Proactively refresh if near expiry
	if time.Until(token.ExpiresAt) <= TokenRefreshGrace {
		logger.GetLogger().Infof("Token near expiry, proactively refreshing for tenant %s", tenantName)
		z.mu.Lock()
		newTokens, err := z.refreshAndSaveToken(ctx, token.RefreshToken, tenantID)
		if err != nil {
			z.mu.Unlock()
			logger.GetLogger().Warnf("Failed to proactively refresh token: %v", err)
		} else {
			// Update cache with new tokens and update local snapshot
			z.setTokenCacheFromResp(tenantID, token, newTokens)
			z.mu.Unlock()
			// Replace local token view for this request
			token.AccessToken = newTokens.AccessToken
			token.RefreshToken = newTokens.RefreshToken
			expiresIn, _ := strconv.Atoi(newTokens.ExpiresIn)
			token.ExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
		}
	}

	// Build a fresh client for this request from the (possibly refreshed) token
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

	// Resolve template ID per-tenant only
	templateID, err := z.resolveTemplateIDFromToken(token, tenantID)
	if err != nil {
		return backoff.Permanent(fmt.Errorf("failed to resolve Zalo template ID: %w", err))
	}

	operation := func() error {
		return z.attemptSendOTP(ctx, cli, receiver, otp, tenantID, templateID)
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

func (z *ZaloProvider) attemptSendOTP(ctx context.Context, cli *client.ZaloClient, receiver, otp string, tenantID uuid.UUID, templateID int) error {
	resp, err := cli.SendOTP(ctx, receiver, otp, templateID)
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
		// Refresh tokens and persist to DB
		newTokens, err := z.refreshAndSaveToken(ctx, refreshToken, tenantID)
		if err != nil {
			return err
		}
		// Update this request's client
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

// getTokenCached returns a decrypted token possibly from cache
func (z *ZaloProvider) getTokenCached(ctx context.Context, tenantID uuid.UUID) (*domain.ZaloToken, error) {
	z.mu.Lock()
	entry, ok := z.tokenCache[tenantID]
	z.mu.Unlock()
	if ok && entry != nil && time.Since(entry.loadedAt) < TokenCacheTTL {
		return entry.token, nil
	}
	// Load fresh
	tok, err := z.getTokenFromDB(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	z.mu.Lock()
	z.tokenCache[tenantID] = &tokenCacheEntry{token: tok, loadedAt: time.Now()}
	z.mu.Unlock()
	return tok, nil
}

// setTokenCacheFromResp updates cache with new token values after refresh.
// If base is provided, it preserves AppID/SecretKey/TemplateID from base.
func (z *ZaloProvider) setTokenCacheFromResp(tenantID uuid.UUID, base *domain.ZaloToken, resp *client.ZaloTokenRefreshResponse) {
	// Build a decrypted token snapshot for cache
	var snapshot *domain.ZaloToken
	if base != nil {
		snapshot = &domain.ZaloToken{
			ID:            base.ID,
			TenantID:      tenantID,
			AppID:         base.AppID,
			SecretKey:     base.SecretKey,
			AccessToken:   resp.AccessToken,
			RefreshToken:  resp.RefreshToken,
			OtpTemplateID: base.OtpTemplateID,
		}
		if secs, err := strconv.Atoi(resp.ExpiresIn); err == nil {
			snapshot.ExpiresAt = time.Now().Add(time.Duration(secs) * time.Second)
		}
		z.mu.Lock()
		z.tokenCache[tenantID] = &tokenCacheEntry{token: snapshot, loadedAt: time.Now()}
		z.mu.Unlock()
	} else {
		// If base unknown, invalidate cache entry
		z.invalidateTokenCache(tenantID)
	}
}

func (z *ZaloProvider) invalidateTokenCache(tenantID uuid.UUID) {
	z.mu.Lock()
	delete(z.tokenCache, tenantID)
	z.mu.Unlock()
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

	// Update cache with new token snapshot
	z.setTokenCacheFromResp(tenantID, existingDecrypted, resp)

	return nil
}

// resolveTemplateIDFromToken requires per-tenant template ID; no global fallback.
func (z *ZaloProvider) resolveTemplateIDFromToken(tok *domain.ZaloToken, tenantID uuid.UUID) (int, error) {
	if tok == nil || tok.OtpTemplateID == "" {
		return 0, fmt.Errorf("no Zalo OTP template ID configured for tenant %s", tenantID)
	}
	tid, convErr := strconv.Atoi(tok.OtpTemplateID)
	if convErr != nil || tid <= 0 {
		return 0, fmt.Errorf("invalid Zalo OTP template ID for tenant %s", tenantID)
	}
	return tid, nil
}

func (z *ZaloProvider) retryWithBackoff(operation func() error) error {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = MaxRetryTime
	b.InitialInterval = 2 * time.Second

	return backoff.Retry(operation, backoff.WithMaxRetries(b, MaxRetries))
}
