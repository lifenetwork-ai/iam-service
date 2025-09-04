package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
)

type IdentityUserUseCase interface {
	ChallengeWithPhone(
		ctx context.Context,
		tenantID uuid.UUID,
		phone string,
	) (*types.IdentityUserChallengeResponse, *errors.DomainError)

	ChallengeWithEmail(
		ctx context.Context,
		tenantID uuid.UUID,
		email string,
	) (*types.IdentityUserChallengeResponse, *errors.DomainError)

	Register(
		ctx context.Context,
		tenantID uuid.UUID,
		lang string,
		email string,
		phone string,
	) (*types.IdentityUserAuthResponse, *errors.DomainError)

	VerifyRegister(
		ctx context.Context,
		tenantID uuid.UUID,
		flowID string,
		code string,
	) (*types.IdentityUserAuthResponse, *errors.DomainError)

	VerifyLogin(
		ctx context.Context,
		tenantID uuid.UUID,
		flowID string,
		code string,
	) (*types.IdentityUserAuthResponse, *errors.DomainError)

	Login(
		ctx context.Context,
		tenantID uuid.UUID,
		username string,
		password string,
	) (*types.IdentityUserAuthResponse, *errors.DomainError)

	Logout(
		ctx context.Context,
		tenantID uuid.UUID,
	) *errors.DomainError

	RefreshToken(
		ctx context.Context,
		tenantID uuid.UUID,
		accessToken string,
		refreshToken string,
	) (*types.IdentityUserAuthResponse, *errors.DomainError)

	Profile(
		ctx context.Context,
		tenantID uuid.UUID,
	) (*types.IdentityUserResponse, *errors.DomainError)

	AddNewIdentifier(
		ctx context.Context,
		tenantID uuid.UUID,
		globalUserID string,
		identifier string,
		identifierType string,
	) (*types.IdentityUserChallengeResponse, *errors.DomainError)

	UpdateIdentifier(
		ctx context.Context,
		globalUserID string,
		tenantID uuid.UUID,
		tenantUserID string,
		identifier string,
		identifierType string,
	) (*types.IdentityUserChallengeResponse, *errors.DomainError)

	CheckIdentifier(
		ctx context.Context,
		tenantID uuid.UUID,
		identifier string,
	) (bool, string, *errors.DomainError)
}
