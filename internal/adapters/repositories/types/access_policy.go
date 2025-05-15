package interfaces

import (
	"context"

	entities "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

type AccessPolicyRepository interface {
	Get(
		ctx context.Context,
		limit int,
		offset int,
		keyword string,
	) ([]entities.AccessPolicy, error)

	GetByID(
		ctx context.Context,
		id string,
	) (*entities.AccessPolicy, error)

	GetByCode(
		ctx context.Context,
		code string,
	) (*entities.AccessPolicy, error)

	Create(
		ctx context.Context,
		entity entities.AccessPolicy,
	) (*entities.AccessPolicy, error)

	Update(
		ctx context.Context,
		entity entities.AccessPolicy,
	) (*entities.AccessPolicy, error)

	SoftDelete(
		ctx context.Context,
		id string,
	) (*entities.AccessPolicy, error)

	Delete(
		ctx context.Context,
		id string,
	) (*entities.AccessPolicy, error)
}
