package usecases

import (
	"time"

	"github.com/genefriendway/human-network-iam/internal/dto"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type identityUseCase struct {
	identityRepository        interfaces.IdentityRepository
}

func NewIdentityUseCase(
	identityRepository interfaces.IdentityRepository,
) interfaces.IdentityUseCase {
	return &identityUseCase{
		identityRepository:        identityRepository,
	}
}

func (u *identityUseCase) ChallengeWithPhone(phone string) (*dto.IdentityChallengeDTO, error) {
	// Check if the account exists
	account, err := u.identityRepository.GetAccountByPhone(phone)
	if err != nil {
		return nil, err
	}

	// Generate a challenge
	challenge := &dto.IdentityChallengeDTO{
		SessionID: account.ID,
		ChallengeAt: time.Now().Format(time.RFC3339),
	}

	return challenge, nil
}

func (u *identityUseCase) ChallengeWithEmail(email string) (*dto.IdentityChallengeDTO, error) {
	// Check if the account exists
	account, err := u.identityRepository.GetAccountByEmail(email)
	if err != nil {
		return nil, err
	}

	// Generate a challenge
	challenge := &dto.IdentityChallengeDTO{
		SessionID: account.ID,
		ChallengeAt: time.Now().Format(time.RFC3339),
	}

	return challenge, nil
}

func (u *identityUseCase) VerifyChallenge(sessionID string, code string) (*dto.IdentityChallengeDTO, error) {
	return &dto.IdentityChallengeDTO{}, nil
}

func (u *identityUseCase) LogInPassword(username string, password string) (*dto.IdentityChallengeDTO, error) {
	return &dto.IdentityChallengeDTO{}, nil
}

func (u *identityUseCase) LogInWithGoogle() (*dto.IdentityChallengeDTO, error) {
	return &dto.IdentityChallengeDTO{}, nil
}

func (u *identityUseCase) LogInWithFacebook() (*dto.IdentityChallengeDTO, error) {
	return &dto.IdentityChallengeDTO{}, nil
}

func (u *identityUseCase) LogInWithApple() (*dto.IdentityChallengeDTO, error) {
	return &dto.IdentityChallengeDTO{}, nil
}

func (u *identityUseCase) LogOut() error {
	return nil
}

