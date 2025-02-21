package interfaces

import (
	"context"

	"github.com/genefriendway/human-network-iam/internal/delivery/dto"
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

	LogInWithGoogle(
		ctx context.Context,
	) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)

	LogInWithFacebook(
		ctx context.Context,
	) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)

	LogInWithApple(
		ctx context.Context,
	) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)

	Register(
		ctx context.Context,
		payload dto.IdentityUserRegisterDTO,
	) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)

	LogIn(
		ctx context.Context,
		username string,
		password string,
	) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)

	LogOut(
		ctx context.Context,
	) *dto.ErrorDTOResponse
}
