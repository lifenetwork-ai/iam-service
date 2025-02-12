package identity_user

import (
	"github.com/genefriendway/human-network-iam/internal/dto"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
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

func (u *userUseCase) ChallengeWithPhone(phone string) (*dto.IdentityUserChallengeDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) ChallengeWithEmail(email string) (*dto.IdentityUserChallengeDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) VerifyChallenge(sessionID string, code string) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) LogInPassword(username string, password string) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) LogInWithGoogle() (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) LogInWithFacebook() (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) LogInWithApple() (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *userUseCase) LogOut() *dto.ErrorDTOResponse {
	return nil
}




