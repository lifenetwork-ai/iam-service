package interfaces

import (
	"context"

	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
)

type PermissionUseCase interface {
	CheckPermission(ctx context.Context, request types.CheckPermissionRequest) (bool, *domainerrors.DomainError)
	DelegateAccess(ctx context.Context, request types.DelegateAccessRequest) (bool, *domainerrors.DomainError)
	CreateRelationTuple(ctx context.Context, request types.CreateRelationTupleRequest) *domainerrors.DomainError
}
