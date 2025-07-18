package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
)

type IdentityUserUseCase interface {
	ChallengeWithPhone(
		ctx context.Context,
		tenantID uuid.UUID,
		phone string,
	) (*dto.IdentityUserChallengeDTO, *errors.DomainError)

	ChallengeWithEmail(
		ctx context.Context,
		tenantID uuid.UUID,
		email string,
	) (*dto.IdentityUserChallengeDTO, *errors.DomainError)

	ChallengeVerify(
		ctx context.Context,
		tenantID uuid.UUID,
		sessionID string,
		code string,
	) (*dto.IdentityUserAuthDTO, *errors.DomainError)

	Register(
		ctx context.Context,
		tenantID uuid.UUID,
		payload dto.IdentityUserRegisterDTO,
	) (*dto.IdentityUserAuthDTO, *errors.DomainError)

	VerifyRegister(
		ctx context.Context,
		tenantID uuid.UUID,
		flowID string,
		code string,
	) (*dto.IdentityUserAuthDTO, *errors.DomainError)

	VerifyLogin(
		ctx context.Context,
		tenantID uuid.UUID,
		flowID string,
		code string,
	) (*dto.IdentityUserAuthDTO, *errors.DomainError)

	LogIn(
		ctx context.Context,
		tenantID uuid.UUID,
		username string,
		password string,
	) (*dto.IdentityUserAuthDTO, *errors.DomainError)

	LogOut(
		ctx context.Context,
		tenantID uuid.UUID,
	) *errors.DomainError

	RefreshToken(
		ctx context.Context,
		tenantID uuid.UUID,
		accessToken string,
		refreshToken string,
	) (*dto.IdentityUserAuthDTO, *errors.DomainError)

	Profile(
		ctx context.Context,
		tenantID uuid.UUID,
	) (*dto.IdentityUserDTO, *errors.DomainError)

	// ChangeIdentifierWithRegisterFlow changes the user's identifier (email or phone)
	ChangeIdentifierWithRegisterFlow(
		ctx context.Context,
		tenantID uuid.UUID,
		newIdentifier string,
	) (*dto.IdentityUserAuthDTO, *errors.DomainError)
}
