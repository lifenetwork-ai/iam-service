package ucases

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms/client"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms/common"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

type smsTokenUseCase struct {
	zaloRepository  domainrepo.ZaloTokenRepository
	zaloTokenCrypto *common.ZaloTokenCrypto
}

func NewSmsTokenUseCase(zaloRepository domainrepo.ZaloTokenRepository, dbEncryptionKey string) interfaces.SmsTokenUseCase {
	return &smsTokenUseCase{
		zaloRepository:  zaloRepository,
		zaloTokenCrypto: common.NewZaloTokenCrypto(dbEncryptionKey),
	}
}

// GetZaloToken retrieves token for a specific tenant
func (u *smsTokenUseCase) GetZaloToken(ctx context.Context, tenantID uuid.UUID) (*domain.ZaloToken, *domainerrors.DomainError) {
	token, err := u.zaloRepository.Get(ctx, tenantID)
	if err != nil {
		return nil, domainerrors.NewInternalError("MSG_GET_TOKEN_FAILED", "Failed to get token")
	}

	decryptedToken, err := u.zaloTokenCrypto.Decrypt(ctx, token)
	if err != nil {
		return nil, domainerrors.NewInternalError("MSG_DECRYPT_TOKEN_FAILED", "Failed to decrypt token")
	}

	return decryptedToken, nil
}

// CreateOrUpdateZaloToken creates/updates Zalo config for a tenant
// If accessToken is empty, automatically refreshes using refreshToken
func (u *smsTokenUseCase) CreateOrUpdateZaloToken(ctx context.Context, tenantID uuid.UUID, appID, secretKey, refreshToken, accessToken string) *domainerrors.DomainError {
	if appID == "" || secretKey == "" || refreshToken == "" {
		return domainerrors.NewValidationError("MSG_INVALID_REQUEST", "app_id, secret_key, and refresh_token are required", nil)
	}

	// If access token not provided or empty, refresh to get a fresh one
	var expiresAt time.Time
	if accessToken == "" {
		logger.GetLogger().Info("Access token not provided, refreshing...")

		// Use Zalo OAuth base URL (hardcoded as it's constant)
		zaloOAuthBaseURL := "https://oauth.zaloapp.com/v4"
		cli, err := client.NewZaloClient(ctx, zaloOAuthBaseURL, secretKey, appID, "", refreshToken)
		if err != nil {
			return domainerrors.WrapInternal(err, "MSG_PROVIDER_BOOTSTRAP_FAIL", "Failed to bootstrap Zalo client")
		}

		resp, err := cli.RefreshAccessToken(ctx, refreshToken)
		if err != nil {
			return domainerrors.WrapInternal(err, "MSG_REFRESH_TOKEN_FAILED", "Failed to refresh Zalo token")
		}

		accessToken = resp.AccessToken
		refreshToken = resp.RefreshToken

		// Convert expires_in string to time
		expiresIn, convErr := strconv.Atoi(resp.ExpiresIn)
		if convErr != nil {
			return domainerrors.WrapInternal(fmt.Errorf("invalid expires_in: %w", convErr), "MSG_REFRESH_TOKEN_FAILED", "Failed to parse expires_in")
		}
		expiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
	} else {
		// If access token provided, assume it expires in 90 days (Zalo default)
		expiresAt = time.Now().Add(90 * 24 * time.Hour)
	}

	// Create/update token
	dbToken := &domain.ZaloToken{
		TenantID:     tenantID,
		AppID:        appID,
		SecretKey:    secretKey,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}

	// Encrypt sensitive fields
	encrypted, err := u.zaloTokenCrypto.Encrypt(ctx, dbToken)
	if err != nil {
		return domainerrors.WrapInternal(err, "MSG_ENCRYPT_TOKEN_FAILED", "Failed to encrypt token")
	}

	// Save to repository
	if err := u.zaloRepository.Save(ctx, encrypted); err != nil {
		return domainerrors.WrapInternal(err, "MSG_SET_TOKEN_FAILED", "Failed to save token")
	}

	return nil
}

// RefreshZaloToken manually refreshes a tenant's token
func (u *smsTokenUseCase) RefreshZaloToken(ctx context.Context, tenantID uuid.UUID, refreshToken string) *domainerrors.DomainError {
	if refreshToken == "" {
		return domainerrors.NewValidationError("MSG_INVALID_REQUEST", "refresh token required", nil)
	}

	// Get existing token to retrieve app_id and secret_key
	existingToken, err := u.zaloRepository.Get(ctx, tenantID)
	if err != nil {
		return domainerrors.WrapInternal(err, "MSG_GET_TOKEN_FAILED", "Failed to get existing token")
	}

	// Decrypt to get credentials
	decrypted, err := u.zaloTokenCrypto.Decrypt(ctx, existingToken)
	if err != nil {
		return domainerrors.WrapInternal(err, "MSG_DECRYPT_TOKEN_FAILED", "Failed to decrypt token")
	}

	// Use Zalo OAuth base URL
	zaloOAuthBaseURL := "https://oauth.zaloapp.com/v4"
	cli, err := client.NewZaloClient(ctx, zaloOAuthBaseURL, decrypted.SecretKey, decrypted.AppID, "", refreshToken)
	if err != nil {
		return domainerrors.WrapInternal(err, "MSG_PROVIDER_BOOTSTRAP_FAIL", "Failed to bootstrap Zalo client")
	}

	resp, err := cli.RefreshAccessToken(ctx, refreshToken)
	if err != nil {
		return domainerrors.WrapInternal(err, "MSG_REFRESH_TOKEN_FAILED", "Failed to refresh Zalo token")
	}

	// Convert expires_in string to time
	expiresIn, convErr := strconv.Atoi(resp.ExpiresIn)
	if convErr != nil {
		return domainerrors.WrapInternal(fmt.Errorf("invalid expires_in: %w", convErr), "MSG_REFRESH_TOKEN_FAILED", "Failed to parse expires_in")
	}

	// Update token
	dbToken := &domain.ZaloToken{
		ID:           existingToken.ID,
		TenantID:     tenantID,
		AppID:        decrypted.AppID,
		SecretKey:    decrypted.SecretKey,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(expiresIn) * time.Second),
	}

	encrypted, encErr := u.zaloTokenCrypto.Encrypt(ctx, dbToken)
	if encErr != nil {
		return domainerrors.WrapInternal(encErr, "MSG_ENCRYPT_TOKEN_FAILED", "Failed to encrypt token")
	}
	if saveErr := u.zaloRepository.Save(ctx, encrypted); saveErr != nil {
		return domainerrors.WrapInternal(saveErr, "MSG_SET_TOKEN_FAILED", "Failed to save refreshed token")
	}

	return nil
}

// DeleteZaloToken removes a tenant's Zalo configuration
func (u *smsTokenUseCase) DeleteZaloToken(ctx context.Context, tenantID uuid.UUID) *domainerrors.DomainError {
	if err := u.zaloRepository.Delete(ctx, tenantID); err != nil {
		return domainerrors.WrapInternal(err, "MSG_DELETE_TOKEN_FAILED", "Failed to delete token")
	}
	return nil
}

// ZaloHealthCheck tests if tenant's Zalo token is valid
func (u *smsTokenUseCase) ZaloHealthCheck(ctx context.Context, tenantID uuid.UUID) *domainerrors.DomainError {
	// Get token from repository
	token, derr := u.GetZaloToken(ctx, tenantID)
	if derr != nil {
		return derr
	}

	// Use Zalo API base URL (hardcoded as it's constant)
	zaloBaseURL := "https://business.openapi.zalo.me"
	cli, err := client.NewZaloClient(ctx, zaloBaseURL, token.SecretKey, token.AppID, token.AccessToken, token.RefreshToken)
	if err != nil {
		return domainerrors.WrapInternal(err, "MSG_CREATE_ZALO_CLIENT_FAILED", "Failed to create Zalo client")
	}
	resp, err := cli.GetAllTemplates(ctx)
	if err != nil {
		return domainerrors.WrapInternal(err, "MSG_GET_ALL_TEMPLATES_FAILED", "Failed to get all templates")
	}

	// Check if error is non-zero - handle both int and float64 from JSON unmarshaling
	var errorCode float64
	switch v := resp.Error.(type) {
	case float64:
		errorCode = v
	case int:
		errorCode = float64(v)
	default:
		return domainerrors.NewInternalError("MSG_GET_ALL_TEMPLATES_FAILED", "Failed to parse error code from response")
	}

	if errorCode != 0 {
		return domainerrors.NewInternalError("MSG_GET_ALL_TEMPLATES_FAILED", "Failed to get all templates")
	}

	return nil
}
