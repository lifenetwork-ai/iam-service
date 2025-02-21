package interfaces

import (
	"context"

	entities "github.com/genefriendway/human-network-iam/internal/domain/entities"
)

type IdentityUserRepository interface {
	GetByPhone(
		ctx context.Context,
		phone string,
	) (*entities.IdentityUser, error)

	GetByEmail(
		ctx context.Context,
		email string,
	) (*entities.IdentityUser, error)

	GetByUsername(
		ctx context.Context,
		username string,
	) (*entities.IdentityUser, error)

	GetByLifeAIID(
		ctx context.Context,
		lifeAIID string,
	) (*entities.IdentityUser, error)

	GetByGoogleID(
		ctx context.Context,
		googleID string,
	) (*entities.IdentityUser, error)

	GetByFacebookID(
		ctx context.Context,
		facebookID string,
	) (*entities.IdentityUser, error)

	GetByAppleID(
		ctx context.Context,
		appleID string,
	) (*entities.IdentityUser, error)

	Create(
		ctx context.Context,
		user *entities.IdentityUser,
	) error

	Update(
		ctx context.Context,
		user *entities.IdentityUser,
	) error

	Delete(
		ctx context.Context,
		userID string,
	) error
}
