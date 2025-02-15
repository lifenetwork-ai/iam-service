package interfaces

import (
	"context"

	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/dto"
)

type IdentityUserUseCase interface {
	ChallengeWithPhone(
		ctx context.Context,
		organizationId string,
		phone string,
	) (*dto.IdentityUserChallengeDTO, *dto.ErrorDTOResponse)

	ChallengeWithEmail(
		ctx context.Context,
		organizationId string,
		email string,
	) (*dto.IdentityUserChallengeDTO, *dto.ErrorDTOResponse)

	ChallengeVerify(
		ctx context.Context,
		sessionID string,
		code string,
	) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)

	LogInWithGoogle(
		ctx context.Context,
		organizationId string,
	) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)

	LogInWithFacebook(
		ctx context.Context,
		organizationId string,
	) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)

	LogInWithApple(
		ctx context.Context,
		organizationId string,
	) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)

	Register(
		ctx context.Context,
		organizationId string,
		payload dto.IdentityUserRegisterDTO,
	) (*dto.IdentityUserAuthDTO, *dto.ErrorDTOResponse)

	LogIn(
		ctx context.Context,
		organizationId string,
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
		organizationId string,
		phone string,
	) (*dto.IdentityUserDTO, error)

	GetByEmail(
		ctx context.Context,
		organizationId string,
		email string,
	) (*dto.IdentityUserDTO, error)

	GetByUsername(
		ctx context.Context,
		organizationId string,
		username string,
	) (*dto.IdentityUserDTO, error)

	GetByGoogleID(
		ctx context.Context,
		organizationId string,
		googleID string,
	) (*dto.IdentityUserDTO, error)

	GetByFacebookID(
		ctx context.Context,
		organizationId string,
		facebookID string,
	) (*dto.IdentityUserDTO, error)

	GetByAppleID(
		ctx context.Context,
		organizationId string,
		appleID string,
	) (*dto.IdentityUserDTO, error)

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
