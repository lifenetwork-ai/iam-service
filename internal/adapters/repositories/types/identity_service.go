package interfaces

import (
	"context"

	entities "github.com/genefriendway/human-network-iam/internal/domain/entities"
)

type IdentityServiceRepository interface {
	Get(
		ctx context.Context,
		limit int,
		offset int,
		keyword string,
	) ([]entities.IdentityService, error)

	GetByID(
		ctx context.Context,
		id string,
	) (*entities.IdentityService, error)

	GetByCode(
		ctx context.Context,
		code string,
	) (*entities.IdentityService, error)

	Create(
		ctx context.Context,
		entity entities.IdentityService,
	) (*entities.IdentityService, error)

	Update(
		ctx context.Context,
		entity entities.IdentityService,
	) (*entities.IdentityService, error)

	SoftDelete(
		ctx context.Context,
		id string,
	) (*entities.IdentityService, error)

	Delete(
		ctx context.Context,
		id string,
	) (*entities.IdentityService, error)
}
