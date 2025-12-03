package interfaces

import (
	"context"

	"github.com/google/uuid"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
)

type SmsTokenUseCase interface {
	// GetZaloToken retrieves token for a specific tenant
	GetZaloToken(ctx context.Context, tenantID uuid.UUID) (*domain.ZaloToken, *domainerrors.DomainError)

	// GetEncryptedZaloToken retrieves token for a specific tenant
	GetEncryptedZaloToken(ctx context.Context, tenantID uuid.UUID) (*domain.ZaloToken, *domainerrors.DomainError)

	// CreateOrUpdateZaloToken creates/updates Zalo config for a tenant
	// If accessToken is empty, automatically refreshes using refreshToken
	CreateOrUpdateZaloToken(ctx context.Context, tenantID uuid.UUID, appID, secretKey, refreshToken, accessToken string) *domainerrors.DomainError

	// RefreshZaloToken manually refreshes a tenant's token
	RefreshZaloToken(ctx context.Context, tenantID uuid.UUID, refreshToken string) *domainerrors.DomainError

	// DeleteZaloToken removes a tenant's Zalo configuration
	DeleteZaloToken(ctx context.Context, tenantID uuid.UUID) *domainerrors.DomainError

	// ZaloHealthCheck tests if tenant's Zalo token is valid
	ZaloHealthCheck(ctx context.Context, tenantID uuid.UUID) *domainerrors.DomainError
}
