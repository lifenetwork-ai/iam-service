package interfaces

import (
	"context"

	entities "github.com/genefriendway/human-network-iam/internal/domain/entities"
)

type AccessPermissionRepository interface {
	Get(
		ctx context.Context,
		limit int,
		offset int,
		keyword string,
	) ([]entities.AccessPermission, error)

	GetByID(
		ctx context.Context,
		id string,
	) (*entities.AccessPermission, error)

	GetByCode(
		ctx context.Context,
		code string,
	) (*entities.AccessPermission, error)

	Create(
		ctx context.Context,
		entity entities.AccessPermission,
	) (*entities.AccessPermission, error)

	Update(
		ctx context.Context,
		entity entities.AccessPermission,
	) (*entities.AccessPermission, error)

	SoftDelete(
		ctx context.Context,
		id string,
	) (*entities.AccessPermission, error)

	Delete(
		ctx context.Context,
		id string,
	) (*entities.AccessPermission, error)
}
