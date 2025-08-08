package common

import (
	"context"

	"github.com/lifenetwork-ai/iam-service/conf"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	"github.com/lifenetwork-ai/iam-service/packages/utils"
)

type ZaloTokenCrypto struct{}

func NewZaloTokenCrypto() *ZaloTokenCrypto {
	return &ZaloTokenCrypto{}
}

func (c *ZaloTokenCrypto) Encrypt(ctx context.Context, token *domain.ZaloToken) (*domain.ZaloToken, error) {
	key := conf.GetDbEncryptionKey()
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
	key := conf.GetDbEncryptionKey()
	decryptedAccess, err := utils.Decrypt(key, token.AccessToken)
	if err != nil {
		return nil, err
	}
	decryptedRefresh, err := utils.Decrypt(key, token.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &domain.ZaloToken{
		AccessToken:  decryptedAccess,
		RefreshToken: decryptedRefresh,
		UpdatedAt:    token.UpdatedAt,
		ExpiresAt:    token.ExpiresAt,
	}, nil
}
