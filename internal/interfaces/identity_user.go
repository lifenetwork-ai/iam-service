package interfaces

import "github.com/genefriendway/human-network-iam/internal/dto"

type IdentityUserUseCase interface {
	ChallengeWithPhone(phone string) (*dto.IdentityUserChallengeDTO, *dto.ErrorDTOResponse)
	ChallengeWithEmail(email string) (*dto.IdentityUserChallengeDTO, *dto.ErrorDTOResponse)
	VerifyChallenge(sessionID string, code string) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)
	LogInPassword(username string, password string) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)
	LogInWithGoogle() (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)
	LogInWithFacebook() (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)
	LogInWithApple() (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)
	LogOut() *dto.ErrorDTOResponse
}

type IdentityUserRepository interface {
	GetByPhone(phone string) (*dto.IdentityUserDTO, error)
	GetByEmail(email string) (*dto.IdentityUserDTO, error)
	GetByUsername(username string) (*dto.IdentityUserDTO, error)
	GetByGoogleID(googleID string) (*dto.IdentityUserDTO, error)
	GetByFacebookID(facebookID string) (*dto.IdentityUserDTO, error)
	GetByAppleID(appleID string) (*dto.IdentityUserDTO, error)
	Create(user *dto.IdentityUserDTO) error
	Update(user *dto.IdentityUserDTO) error
	Delete(userID string) error
}