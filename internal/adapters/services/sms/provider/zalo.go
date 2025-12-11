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
	MaxRetryTime      = 30 * time.Second
	MaxRetries        = 3
	TokenInvalidError = -124
	SuccessCode       = 0
	// TokenCacheTTL defines how long a decrypted token stays in cache
	TokenCacheTTL = 30 * time.Second
)

// ZaloProvider handles messages through Zalo in multi-tenant mode.
// Clients are constructed per request. A small cache keeps decrypted tokens to
// avoid repeated decrypt/DB hits. No per-request refresh is performed; tokens
// are refreshed exclusively by the background worker.
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
func (z *ZaloProvider) SendOTP(ctx context.Context, tenantName, receiver, otp string, ttl time.Duration, _ string) error {
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

	// Strict worker refresh mode: do not proactively refresh here. The background worker handles token refresh.

	// Build a fresh client for this request from the token
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
		return z.attemptSendOTP(ctx, cli, receiver, otp, templateID)
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

func (z *ZaloProvider) attemptSendOTP(ctx context.Context, cli *client.ZaloClient, receiver, otp string, templateID int) error {
	resp, err := cli.SendOTP(ctx, receiver, otp, templateID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to send OTP via Zalo: %v", err)
		return fmt.Errorf("failed to send OTP via Zalo: %w", err)
	}

	return z.handleAPIResponse(resp)
}

func (z *ZaloProvider) handleAPIResponse(resp *client.ZaloSendNotificationResponse) error {
	switch resp.Error {
	case SuccessCode:
		return nil
	case TokenInvalidError:
		logger.GetLogger().Warn("Access token invalid; strict worker refresh mode enabled. No reactive refresh in provider.")
		return backoff.Permanent(fmt.Errorf("zalo access token invalid; background refresh required"))
	default:
		return backoff.Permanent(fmt.Errorf("zalo api error: code %d, message: %s", resp.Error, resp.Message))
	}
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
