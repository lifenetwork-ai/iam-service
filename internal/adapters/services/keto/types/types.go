package types

import (
	"context"

	ucasetypes "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
)

type KetoService interface {
	CheckPermission(ctx context.Context, request ucasetypes.CheckPermissionRequest) (bool, error)
	// BatchCheckPermission(ctx context.Context, dto dto.BatchCheckPermissionRequestDTO) (bool, error)
	CreateRelationTuple(ctx context.Context, request ucasetypes.CreateRelationTupleRequest) error
}
