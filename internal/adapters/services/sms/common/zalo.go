package common

import (
	"context"

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

func (c *ZaloTokenCrypto) Encrypt(ctx context.Context, token *domain.ZaloToken) (*domain.ZaloToken, error) {
	key := [32]byte{}
	copy(key[:], []byte(c.dbEncryptionKey))
	encryptedAccess, err := utils.Encrypt(key, token.AccessToken)
	if err != nil {
		return nil, err
	}
	encryptedRefresh, err := utils.Encrypt(key, token.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &domain.ZaloToken{
		AccessToken:  encryptedAccess,
		RefreshToken: encryptedRefresh,
		UpdatedAt:    token.UpdatedAt,
		ExpiresAt:    token.ExpiresAt,
	}, nil
}

func (c *ZaloTokenCrypto) Decrypt(ctx context.Context, token *domain.ZaloToken) (*domain.ZaloToken, error) {
	key := [32]byte{}
	copy(key[:], []byte(c.dbEncryptionKey))
	decryptedAccess, err := utils.Decrypt(key, token.AccessToken)
	if err != nil {
		return nil, err
	}
	decryptedRefresh, err := utils.Decrypt(key, token.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &domain.ZaloToken{
		ID:           token.ID,
		AccessToken:  decryptedAccess,
		RefreshToken: decryptedRefresh,
		UpdatedAt:    token.UpdatedAt,
		ExpiresAt:    token.ExpiresAt,
	}, nil
}
