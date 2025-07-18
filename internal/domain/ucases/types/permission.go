package interfaces

import (
	"context"

	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
)

type PermissionUseCase interface {
	CheckPermission(ctx context.Context, dto dto.CheckPermissionRequestDTO) (bool, *domainerrors.DomainError)
	BatchCheckPermission(ctx context.Context, dto dto.BatchCheckPermissionRequestDTO) (bool, *domainerrors.DomainError)
	CreateRelationTuple(ctx context.Context, dto dto.CreateRelationTupleRequestDTO) *domainerrors.DomainError
}
