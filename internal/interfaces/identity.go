package interfaces

import (
	"github.com/genefriendway/human-network-iam/internal/dto"
)

type IdentityUseCase interface {
	ChallengeWithPhone(phone string) (*dto.IdentityChallengeDTO, error)
	ChallengeWithEmail(email string) (*dto.IdentityChallengeDTO, error)
	VerifyChallenge(sessionID string, code string) (*dto.IdentityChallengeDTO, error)
	LogInPassword(username string, password string) (*dto.IdentityChallengeDTO, error)
	LogInWithGoogle() (*dto.IdentityChallengeDTO, error)
	LogInWithFacebook() (*dto.IdentityChallengeDTO, error)
	LogInWithApple() (*dto.IdentityChallengeDTO, error)
	LogOut() error
}

type IdentityRepository interface {
	GetAccountByPhone(phone string) (*dto.AccountDTO, error)
	GetAccountByEmail(email string) (*dto.AccountDTO, error)
	GetAccountByUsername(username string) (*dto.AccountDTO, error)
	GetAccountByGoogleID(googleID string) (*dto.AccountDTO, error)
	GetAccountByFacebookID(facebookID string) (*dto.AccountDTO, error)
	GetAccountByAppleID(appleID string) (*dto.AccountDTO, error)
	CreateAccount(user *dto.AccountDTO) error
	UpdateAccount(user *dto.AccountDTO) error
	DeleteAccount(userID string) error
}
