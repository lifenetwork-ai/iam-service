package ucases

import (
	"context"

	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
)

type smsTokenUseCase struct {
	zaloRepository domainrepo.ZaloTokenRepository
}

func NewSmsTokenUseCase(zaloRepository domainrepo.ZaloTokenRepository) interfaces.SmsTokenUseCase {
	return &smsTokenUseCase{
		zaloRepository: zaloRepository,
	}
}

// GetToken gets the token from the repository
// Should only be called by admin or authorized user
func (u *smsTokenUseCase) GetZaloToken(ctx context.Context) (*domain.ZaloToken, *domainerrors.DomainError) {
	token, err := u.zaloRepository.Get(ctx)
	if err != nil {
		return nil, domainerrors.NewInternalError("MSG_GET_TOKEN_FAILED", "Failed to get token")
	}

	return token, nil
}

// SetToken sets the token in the repository
// Should only be called by admin or authorized user
func (u *smsTokenUseCase) SetZaloToken(ctx context.Context, accessToken, refreshToken string) *domainerrors.DomainError {
	token := &domain.ZaloToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	err := u.zaloRepository.Save(ctx, token)
	if err != nil {
		return domainerrors.NewInternalError("MSG_SET_TOKEN_FAILED", "Failed to set token")
	}

	return nil
}
