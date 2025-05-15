package interfaces

import (
	"context"

	entities "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

type IdentityRoleRepository interface {
	Get(
		ctx context.Context,
		limit int,
		offset int,
		keyword string,
	) ([]entities.IdentityRole, error)

	GetByID(
		ctx context.Context,
		id string,
	) (*entities.IdentityRole, error)

	GetByCode(
		ctx context.Context,
		code string,
	) (*entities.IdentityRole, error)

	Create(
		ctx context.Context,
		entity entities.IdentityRole,
	) (*entities.IdentityRole, error)

	Update(
		ctx context.Context,
		entity entities.IdentityRole,
	) (*entities.IdentityRole, error)

	SoftDelete(
		ctx context.Context,
		id string,
	) (*entities.IdentityRole, error)

	Delete(
		ctx context.Context,
		id string,
	) (*entities.IdentityRole, error)
}
