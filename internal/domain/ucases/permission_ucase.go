package ucases

import (
	"context"

	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

type permissionUseCase struct {
	ketoClient KetoService
}

func NewPermissionUseCase(ketoClient KetoService) interfaces.PermissionUseCase {
	return &permissionUseCase{
		ketoClient: ketoClient,
	}
}

// CheckPermission checks if a subject has permission to perform an action on an object.
// This function will be used to check if a subject has permission to perform an action on an object.
// Should only be used internally to check if a subject has permission to perform an action on an object.
func (u *permissionUseCase) CheckPermission(ctx context.Context, request types.CheckPermissionRequest) (bool, *domainerrors.DomainError) {
	if err := request.Validate(); err != nil {
		logger.GetLogger().Errorf("Invalid check permission request: %v", err)
		return false, domainerrors.NewValidationError(
			"MSG_INVALID_CHECK_PERMISSION_REQUEST",
			"Invalid check permission request",
			err,
		)
	}

	allowed, err := u.ketoClient.CheckPermission(ctx, request)
	if err != nil {
		logger.GetLogger().Errorf("Failed to check permission: %v", err)
		return false, domainerrors.NewInternalError(
			"MSG_CHECK_PERMISSION_FAILED",
			"Failed to check permission",
		)
	}

	return allowed, nil
}

// func (u *permissionUseCase) BatchCheckPermission(ctx context.Context, dto dto.BatchCheckPermissionRequestDTO) (bool, *domainerrors.DomainError) {
// 	allowed, err := u.ketoClient.BatchCheckPermission(ctx, dto)
// 	if err != nil {
// 		logger.GetLogger().Errorf("Failed to batch check permission: %v", err)
// 		return false, domainerrors.NewInternalError(
// 			"MSG_BATCH_CHECK_PERMISSION_FAILED",
// 			"Failed to batch check permission",
// 		)
// 	}

// 	return allowed, nil
// }

// CreateRelationTuple creates a relation tuple
// This function will be used to create a relation tuple for a tenant member
// Should only be used internally to create a relation tuple for a tenant member
func (u *permissionUseCase) CreateRelationTuple(ctx context.Context, request types.CreateRelationTupleRequest) *domainerrors.DomainError {
	if err := request.Validate(); err != nil {
		logger.GetLogger().Errorf("Invalid create relation tuple request: %v", err)
		return domainerrors.NewValidationError(
			"MSG_INVALID_CREATE_RELATION_TUPLE_REQUEST",
			"Invalid create relation tuple request",
			err,
		)
	}

	err := u.ketoClient.CreateRelationTuple(ctx, request)
	if err != nil {
		logger.GetLogger().Errorf("Failed to create relation tuple: %v", err)
		return domainerrors.NewInternalError(
			"MSG_CREATE_RELATION_TUPLE_FAILED",
			"Failed to create relation tuple",
		)
	}

	return nil
}
