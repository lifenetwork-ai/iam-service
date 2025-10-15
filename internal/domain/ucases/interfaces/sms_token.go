package interfaces

import (
	"context"

	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
)

type SmsTokenUseCase interface {
	// GetToken gets the token from the repository
	GetZaloToken(ctx context.Context) (*domain.ZaloToken, *domainerrors.DomainError)

	// SetToken sets the token in the repository
	SetZaloToken(ctx context.Context, accessToken, refreshToken string) *domainerrors.DomainError

	// RefreshZaloToken refreshes tokens via admin-provided refresh token
	RefreshZaloToken(ctx context.Context, refreshToken string) *domainerrors.DomainError

	// Zalo health check
	ZaloHealthCheck(ctx context.Context) *domainerrors.DomainError
}
