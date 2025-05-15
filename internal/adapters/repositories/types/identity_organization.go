package interfaces

import (
	"context"

	entities "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

type IdentityOrganizationRepository interface {
	Get(
		ctx context.Context,
		limit int,
		offset int,
		keyword string,
	) ([]entities.IdentityOrganization, error)

	GetByID(
		ctx context.Context,
		id string,
	) (*entities.IdentityOrganization, error)

	GetByCode(
		ctx context.Context,
		code string,
	) (*entities.IdentityOrganization, error)

	Create(
		ctx context.Context,
		entity entities.IdentityOrganization,
	) (*entities.IdentityOrganization, error)

	Update(
		ctx context.Context,
		entity entities.IdentityOrganization,
	) (*entities.IdentityOrganization, error)

	SoftDelete(
		ctx context.Context,
		id string,
	) (*entities.IdentityOrganization, error)

	Delete(
		ctx context.Context,
		id string,
	) (*entities.IdentityOrganization, error)
}
