package types

import (
	"context"

	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
)

type KetoService interface {
	CheckPermission(ctx context.Context, dto dto.CheckPermissionRequestDTO) (bool, error)
	BatchCheckPermission(ctx context.Context, dto dto.BatchCheckPermissionRequestDTO) (bool, error)
	CreateRelationTuple(ctx context.Context, dto dto.CreateRelationTupleRequestDTO) error
}
