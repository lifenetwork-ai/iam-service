package ucases

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms/client"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms/common"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
)

type smsTokenUseCase struct {
	zaloRepository  domainrepo.ZaloTokenRepository
	zaloTokenCrypto *common.ZaloTokenCrypto
}

func NewSmsTokenUseCase(zaloRepository domainrepo.ZaloTokenRepository) interfaces.SmsTokenUseCase {
	return &smsTokenUseCase{
		zaloRepository:  zaloRepository,
		zaloTokenCrypto: common.NewZaloTokenCrypto(conf.GetConfiguration().DbEncryptionKey),
	}
}

// GetToken gets the token from the repository
func (u *smsTokenUseCase) GetZaloToken(ctx context.Context) (*domain.ZaloToken, *domainerrors.DomainError) {
	token, err := u.zaloRepository.Get(ctx)
	if err != nil {
		return nil, domainerrors.NewInternalError("MSG_GET_TOKEN_FAILED", "Failed to get token")
	}

	decryptedToken, err := u.zaloTokenCrypto.Decrypt(ctx, token)
	if err != nil {
		return nil, domainerrors.NewInternalError("MSG_DECRYPT_TOKEN_FAILED", "Failed to decrypt token")
	}

	return decryptedToken, nil
}

// SetToken sets the token in the repository
func (u *smsTokenUseCase) SetZaloToken(ctx context.Context, accessToken, refreshToken string) *domainerrors.DomainError {
	encryptedToken, err := u.zaloTokenCrypto.Encrypt(ctx, &domain.ZaloToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
	if err != nil {
		return domainerrors.NewInternalError("MSG_ENCRYPT_TOKEN_FAILED", "Failed to encrypt token")
	}

	err = u.zaloRepository.Save(ctx, encryptedToken)
	if err != nil {
		return domainerrors.NewInternalError("MSG_SET_TOKEN_FAILED", "Failed to set token")
	}

	return nil
}

// RefreshZaloToken refreshes and persists Zalo tokens using an admin-provided refresh token.
// This bypasses any invalid or expired state in the DB by bootstrapping a minimal client for refresh.
func (u *smsTokenUseCase) RefreshZaloToken(ctx context.Context, refreshToken string) *domainerrors.DomainError {
	if refreshToken == "" {
		return domainerrors.NewValidationError("MSG_INVALID_REQUEST", "refresh token required", nil)
	}

	cfg := conf.GetConfiguration().Sms.Zalo
	cli, err := client.NewZaloClient(ctx, cfg.ZaloBaseURL, cfg.ZaloSecretKey, cfg.ZaloAppID, "", refreshToken)
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

	// Persist encrypted token to repository
	dbToken := &domain.ZaloToken{
		ID:           1,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(expiresIn) * time.Second),
		UpdatedAt:    time.Now(),
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

// Zalo health check
func (u *smsTokenUseCase) ZaloHealthCheck(ctx context.Context) *domainerrors.DomainError {
	// Get token from repository
	token, derr := u.GetZaloToken(ctx)
	if derr != nil {
		return derr
	}

	cfg := conf.GetConfiguration().Sms.Zalo
	cli, err := client.NewZaloClient(ctx, cfg.ZaloBaseURL, cfg.ZaloSecretKey, cfg.ZaloAppID, token.AccessToken, token.RefreshToken)
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
