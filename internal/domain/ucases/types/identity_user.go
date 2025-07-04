package interfaces

import (
	"context"

	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
)

type IdentityUserUseCase interface {
	ChallengeWithPhone(
		ctx context.Context,
		phone string,
	) (*dto.IdentityUserChallengeDTO, *dto.ErrorDTOResponse)

	ChallengeWithEmail(
		ctx context.Context,
		email string,
	) (*dto.IdentityUserChallengeDTO, *dto.ErrorDTOResponse)

	ChallengeVerify(
		ctx context.Context,
		sessionID string,
		code string,
	) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)

	Register(
		ctx context.Context,
		payload dto.IdentityUserRegisterDTO,
	) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)

	VerifyRegister(
		ctx context.Context,
		flowID string,
		code string,
	) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)

	VerifyLogin(
		ctx context.Context,
		flowID string,
		code string,
	) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)

	LogIn(
		ctx context.Context,
		username string,
		password string,
	) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)

	LogOut(
		ctx context.Context,
	) *dto.ErrorDTOResponse

	RefreshToken(
		ctx context.Context,
		accessToken string,
		refreshToken string,
	) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)

	Profile(
		ctx context.Context,
	) (*dto.IdentityUserDTO, *dto.ErrorDTOResponse)
}
