package interfaces

import (
	"context"

	entities "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

type IdentityUserRepository interface {
	FindByID(
		ctx context.Context,
		userID string,
	) (*entities.IdentityUser, error)

	FindByPhone(
		ctx context.Context,
		phone string,
	) (*entities.IdentityUser, error)

	FindByEmail(
		ctx context.Context,
		email string,
	) (*entities.IdentityUser, error)

	FindByUsername(
		ctx context.Context,
		username string,
	) (*entities.IdentityUser, error)

	FindBySelfAuthenticateID(
		ctx context.Context,
		selfAuthID string,
	) (*entities.IdentityUser, error)

	FindByGoogleID(
		ctx context.Context,
		googleID string,
	) (*entities.IdentityUser, error)

	FindByFacebookID(
		ctx context.Context,
		facebookID string,
	) (*entities.IdentityUser, error)

	FindByAppleID(
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
