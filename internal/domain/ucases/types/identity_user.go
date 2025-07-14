package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
)

type IdentityUserUseCase interface {
	ChallengeWithPhone(
		ctx context.Context,
		tenantID uuid.UUID,
		phone string,
	) (*dto.IdentityUserChallengeDTO, *dto.ErrorDTOResponse)

	ChallengeWithEmail(
		ctx context.Context,
		tenantID uuid.UUID,
		email string,
	) (*dto.IdentityUserChallengeDTO, *dto.ErrorDTOResponse)

	ChallengeVerify(
		ctx context.Context,
		tenantID uuid.UUID,
		sessionID string,
		code string,
	) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)

	Register(
		ctx context.Context,
		tenantID uuid.UUID,
		payload dto.IdentityUserRegisterDTO,
	) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)

	VerifyRegister(
		ctx context.Context,
		tenantID uuid.UUID,
		flowID string,
		code string,
	) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)

	VerifyLogin(
		ctx context.Context,
		tenantID uuid.UUID,
		flowID string,
		code string,
	) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)

	LogIn(
		ctx context.Context,
		tenantID uuid.UUID,
		username string,
		password string,
	) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)

	LogOut(
		ctx context.Context,
		tenantID uuid.UUID,
	) *dto.ErrorDTOResponse

	RefreshToken(
		ctx context.Context,
		tenantID uuid.UUID,
		accessToken string,
		refreshToken string,
	) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)

	Profile(
		ctx context.Context,
		tenantID uuid.UUID,
	) (*dto.IdentityUserDTO, *dto.ErrorDTOResponse)
}
