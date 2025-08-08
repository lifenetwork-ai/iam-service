package ucases

import (
	"context"

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
		zaloTokenCrypto: common.NewZaloTokenCrypto(),
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
