package interfaces

import (
	"context"

	entities "github.com/genefriendway/human-network-iam/internal/domain/entities"
)

type AccessSessionRepository interface {
	Get(
		ctx context.Context,
		limit int,
		offset int,
		keyword string,
	) ([]entities.AccessSession, error)

	FindByID(
		ctx context.Context,
		id string,
	) (*entities.AccessSession, error)

	FindByAccessToken(
		ctx context.Context,
		accessToken string,
	) (*entities.AccessSession, error)

	FindByRefreshToken(
		ctx context.Context,
		refreshToken string,
	) (*entities.AccessSession, error)

	Create(
		ctx context.Context,
		entity *entities.AccessSession,
	) (*entities.AccessSession, error)

	Update(
		ctx context.Context,
		entity *entities.AccessSession,
	) (*entities.AccessSession, error)

	SoftDelete(
		ctx context.Context,
		id string,
	) (*entities.AccessSession, error)

	Delete(
		ctx context.Context,
		id string,
	) (*entities.AccessSession, error)
}
