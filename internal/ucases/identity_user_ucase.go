package ucases

import (
	"context"

	"github.com/genefriendway/human-network-iam/internal/dto"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
	"github.com/genefriendway/human-network-iam/packages/utils"
)

type userUseCase struct {
	userRepo interfaces.IdentityUserRepository
}

func NewIdentityUserUseCase(
	userRepo interfaces.IdentityUserRepository,
) interfaces.IdentityUserUseCase {
	return &userUseCase{
		userRepo: userRepo,
	}
}

func (u *userUseCase) ChallengeWithPhone(
	ctx context.Context,
	organizationId string,
	phone string,
) (*dto.IdentityUserChallengeDTO, *dto.ErrorDTOResponse) {
	if utils.IsPhoneNumber(phone) {
		return nil, &dto.ErrorDTOResponse{
			Code:    "INVALID_PHONE_NUMBER",
			Message: "Invalid phone number",
			Details: nil,
		}
	}
	return nil, nil
}

func (u *userUseCase) ChallengeWithEmail(
	ctx context.Context,
	organizationId string,
	email string,
) (*dto.IdentityUserChallengeDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) ChallengeVerify(
	ctx context.Context,
	sessionID string,
	code string,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) LogInWithGoogle(
	ctx context.Context,
	organizationId string,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) LogInWithFacebook(
	ctx context.Context,
	organizationId string,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) LogInWithApple(
	ctx context.Context,
	organizationId string,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) Register(
	ctx context.Context,
	organizationId string,
	payload dto.IdentityUserRegisterDTO,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) LogIn(
	ctx context.Context,
	organizationId string,
	username string,
	password string,
) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) LogOut(
	ctx context.Context,
) *dto.ErrorDTOResponse {
	return nil
}
