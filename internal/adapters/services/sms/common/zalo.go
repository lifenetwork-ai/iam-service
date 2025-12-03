package common

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"

	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	"github.com/lifenetwork-ai/iam-service/packages/utils"
)

type ZaloTokenCrypto struct {
	dbEncryptionKey string
}

func NewZaloTokenCrypto(dbEncryptionKey string) *ZaloTokenCrypto {
	return &ZaloTokenCrypto{
		dbEncryptionKey: dbEncryptionKey,
	}
}

// deriveKey deterministically derives a 32-byte key from the configured dbEncryptionKey.
// This avoids the previous behavior where short keys were silently zero-padded, which weakens security.
func (c *ZaloTokenCrypto) deriveKey() ([32]byte, error) {
	if c.dbEncryptionKey == "" {
		return [32]byte{}, errors.New("db encryption key is empty")
	}
	// SHA-256 returns a 32-byte array which matches the required key size.
	sum := sha256.Sum256([]byte(c.dbEncryptionKey))
	return sum, nil
}

func (c *ZaloTokenCrypto) Encrypt(ctx context.Context, token *domain.ZaloToken) (*domain.ZaloToken, error) {
	_ = ctx // context reserved for future use (cancellation, tracing)
	if token == nil {
		return nil, errors.New("nil token")
	}

	key, err := c.deriveKey()
	if err != nil {
		return nil, err
	}

	encryptedAccess, err := utils.Encrypt(key, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("encrypt access token: %w", err)
	}
	encryptedRefresh, err := utils.Encrypt(key, token.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("encrypt refresh token: %w", err)
	}
	encryptedSecret, err := utils.Encrypt(key, token.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt secret key: %w", err)
	}

	return &domain.ZaloToken{
		ID:            token.ID,
		AccessToken:   encryptedAccess,
		RefreshToken:  encryptedRefresh,
		SecretKey:     encryptedSecret,
		AppID:         token.AppID,
		TenantID:      token.TenantID,
		OtpTemplateID: token.OtpTemplateID,
		UpdatedAt:     token.UpdatedAt,
		ExpiresAt:     token.ExpiresAt,
		CreatedAt:     token.CreatedAt,
	}, nil
}

func (c *ZaloTokenCrypto) Decrypt(ctx context.Context, token *domain.ZaloToken) (*domain.ZaloToken, error) {
	_ = ctx // context reserved for future use (cancellation, tracing)
	if token == nil {
		return nil, errors.New("nil token")
	}

	key, err := c.deriveKey()
	if err != nil {
		return nil, err
	}

	decryptedAccess, err := utils.Decrypt(key, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("decrypt access token: %w", err)
	}
	decryptedRefresh, err := utils.Decrypt(key, token.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("decrypt refresh token: %w", err)
	}
	decryptedSecret, err := utils.Decrypt(key, token.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("decrypt secret key: %w", err)
	}

	return &domain.ZaloToken{
		ID:            token.ID,
		AccessToken:   decryptedAccess,
		RefreshToken:  decryptedRefresh,
		SecretKey:     decryptedSecret,
		AppID:         token.AppID,
		TenantID:      token.TenantID,
		OtpTemplateID: token.OtpTemplateID,
		UpdatedAt:     token.UpdatedAt,
		ExpiresAt:     token.ExpiresAt,
		CreatedAt:     token.CreatedAt,
	}, nil
}
