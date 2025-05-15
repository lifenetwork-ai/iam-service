package interfaces

import (
	"context"

	entities "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

type IdentityGroupRepository interface {
	Get(
		ctx context.Context,
		limit int,
		offset int,
		keyword string,
	) ([]entities.IdentityGroup, error)

	GetByID(
		ctx context.Context,
		id string,
	) (*entities.IdentityGroup, error)

	GetByCode(
		ctx context.Context,
		code string,
	) (*entities.IdentityGroup, error)

	Create(
		ctx context.Context,
		entity entities.IdentityGroup,
	) (*entities.IdentityGroup, error)

	Update(
		ctx context.Context,
		entity entities.IdentityGroup,
	) (*entities.IdentityGroup, error)

	SoftDelete(
		ctx context.Context,
		id string,
	) (*entities.IdentityGroup, error)

	Delete(
		ctx context.Context,
		id string,
	) (*entities.IdentityGroup, error)
}
