package interfaces

import (
	"context"

	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
)

type AuthorizationService interface {
	CheckPermission(ctx context.Context, request types.CheckPermissionRequest) (bool, error)
	// BatchCheckPermission(ctx context.Context, dto dto.BatchCheckPermissionRequestDTO) (bool, error)
	CreateRelationTuple(ctx context.Context, request types.CreateRelationTupleRequest) error
}
