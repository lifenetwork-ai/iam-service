package workers

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/lifenetwork-ai/iam-service/constants"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms/client"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms/common"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
	"github.com/lifenetwork-ai/iam-service/internal/workers/types"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

type zaloRefreshTokenWorker struct {
	zaloTokenRepo   domainrepo.ZaloTokenRepository
	zaloTokenCrypto *common.ZaloTokenCrypto

	mu      sync.Mutex
	running bool
}

// NewZaloRefreshTokenWorker creates a new worker instance
func NewZaloRefreshTokenWorker(
	zaloTokenRepo domainrepo.ZaloTokenRepository,
	dbEncryptionKey string,
) types.Worker {
	return &zaloRefreshTokenWorker{
		zaloTokenRepo:   zaloTokenRepo,
		zaloTokenCrypto: common.NewZaloTokenCrypto(dbEncryptionKey),
	}
}

// Name returns the worker name
func (w *zaloRefreshTokenWorker) Name() string {
	return "zalo-refresh-token-worker"
}

// Start periodically retries failed OTP deliveries
func (w *zaloRefreshTokenWorker) Start(ctx context.Context, interval time.Duration) {
	logger.GetLogger().Infof("[%s] started with interval %s", w.Name(), interval.String())
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go w.safeProcess(ctx)
		case <-ctx.Done():
			logger.GetLogger().Infof("[%s] stopped", w.Name())
			return
		}
	}
}

// safeProcess checks and prevents concurrent execution
func (w *zaloRefreshTokenWorker) safeProcess(ctx context.Context) {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		logger.GetLogger().Warnf("[%s] still processing, skipping this tick", w.Name())
		return
	}
	w.running = true
	w.mu.Unlock()

	defer func() {
		w.mu.Lock()
		w.running = false
		w.mu.Unlock()
	}()

	w.processZaloToken(ctx)
}

func (w *zaloRefreshTokenWorker) processZaloToken(ctx context.Context) {
	// Fetch all tokens and refresh only those expiring within the next 24 hours
	tokens, err := w.zaloTokenRepo.GetAll(ctx)
	if err != nil {
		logger.GetLogger().Errorf("[%s] failed to fetch tokens for refresh: %v", w.Name(), err)
		return
	}

	if len(tokens) == 0 {
		logger.GetLogger().Infof("[%s] no tokens to evaluate for refresh", w.Name())
		return
	}

	// Refresh all tokens on each tick to ensure periodic rotation regardless of expiry
	toRefresh := tokens
	logger.GetLogger().Infof("[%s] refreshing %d token(s) this tick", w.Name(), len(toRefresh))

	// Refresh each tenant's token
	for _, token := range toRefresh {
		if err := w.refreshTokenForTenant(ctx, token); err != nil {
			logger.GetLogger().Errorf("[%s] failed to refresh token for tenant %s: %v", w.Name(), token.TenantID, err)
			// Continue with other tenants even if one fails
			continue
		}
		logger.GetLogger().Infof("[%s] successfully refreshed token for tenant %s", w.Name(), token.TenantID)
	}
}

func (w *zaloRefreshTokenWorker) refreshTokenForTenant(ctx context.Context, token *domain.ZaloToken) error {
	// Decrypt token to get credentials
	decrypted, err := w.zaloTokenCrypto.Decrypt(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to decrypt token: %w", err)
	}

	// Use Zalo OAuth base URL
	zaloOAuthBaseURL := constants.ZaloOAuthBaseURL
	cli, err := client.NewZaloClient(ctx, zaloOAuthBaseURL, decrypted.SecretKey, decrypted.AppID, "", decrypted.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to create Zalo client: %w", err)
	}

	// Call Zalo API to refresh
	resp, err := cli.RefreshAccessToken(ctx, decrypted.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to refresh access token: %w", err)
	}

	// Convert expires_in string to time
	expiresIn, convErr := strconv.Atoi(resp.ExpiresIn)
	if convErr != nil {
		return fmt.Errorf("invalid expires_in: %w", convErr)
	}

	// Update token with new values
	updatedToken := &domain.ZaloToken{
		ID:           token.ID,
		TenantID:     token.TenantID,
		AppID:        decrypted.AppID,
		SecretKey:    decrypted.SecretKey,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(expiresIn) * time.Second),
	}

	// Encrypt and save
	encrypted, err := w.zaloTokenCrypto.Encrypt(ctx, updatedToken)
	if err != nil {
		return fmt.Errorf("failed to encrypt token: %w", err)
	}

	if err := w.zaloTokenRepo.Save(ctx, encrypted); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	return nil
}
