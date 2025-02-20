package interfaces

import (
	"context"

	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/dto"
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

type IdentityUserRepository interface {
	GetByPhone(
		ctx context.Context,
		phone string,
	) (*domain.IdentityUser, error)

	GetByEmail(
		ctx context.Context,
		email string,
	) (*domain.IdentityUser, error)

	GetByUsername(
		ctx context.Context,
		username string,
	) (*domain.IdentityUser, error)

	GetByLifeAIID(
		ctx context.Context,
		lifeAIID string,
	) (*domain.IdentityUser, error)

	GetByGoogleID(
		ctx context.Context,
		googleID string,
	) (*domain.IdentityUser, error)

	GetByFacebookID(
		ctx context.Context,
		facebookID string,
	) (*domain.IdentityUser, error)

	GetByAppleID(
		ctx context.Context,
		appleID string,
	) (*domain.IdentityUser, error)

	Create(
		ctx context.Context,
		user *domain.IdentityUser,
	) error

	Update(
		ctx context.Context,
		user *domain.IdentityUser,
	) error

	Delete(
		ctx context.Context,
		userID string,
	) error
}
