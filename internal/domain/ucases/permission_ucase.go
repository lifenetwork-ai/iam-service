package ucases

import (
	"context"
	"fmt"

	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
	domainservice "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/services"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
	"github.com/lifenetwork-ai/iam-service/packages/utils"
)

type permissionUseCase struct {
	ketoClient       domainservice.KetoService
	userIdentityRepo domainrepo.UserIdentityRepository
}

func NewPermissionUseCase(ketoClient domainservice.KetoService) interfaces.PermissionUseCase {
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

	globalUserID, err := u.getGlobalUserID(ctx, &request)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get global user id: %v", err)
		return false, domainerrors.NewInternalError(
			"MSG_GET_GLOBAL_USER_ID_FAILED",
			"Failed to get global user id",
		)

	}

	// Set the global user id to the request
	request.GlobalUserID = globalUserID

	allowed, ketoErr := u.ketoClient.CheckPermission(ctx, request)
	if ketoErr != nil {
		logger.GetLogger().Errorf("Failed to check permission: %v", ketoErr)
		return false, domainerrors.NewInternalError(
			"MSG_CHECK_PERMISSION_FAILED",
			"Failed to check permission",
		)
	}

	return allowed, nil
}

func (u *permissionUseCase) DelegateAccess(ctx context.Context, request types.DelegateAccessRequest) (bool, *domainerrors.DomainError) {
	if err := request.Validate(); err != nil {
		logger.GetLogger().Errorf("Invalid delegate access request: %v", err)
		return false, domainerrors.NewValidationError(
			"MSG_INVALID_DELEGATE_ACCESS_REQUEST",
			"Invalid delegate access request",
			err,
		)
	}

	globalUserID, err := u.getGlobalUserID(ctx, &request)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get global user id: %v", err)
		return false, domainerrors.NewInternalError(
			"MSG_GET_GLOBAL_USER_ID_FAILED",
			"Failed to get global user id",
		)
	}

	createRelationTupleRequest := types.CreateRelationTupleRequest{
		Namespace: request.ResourceType,
		Relation:  request.Permission,
		Object:    fmt.Sprintf("%s:%s", request.ResourceType, request.ResourceID),
		TenantRelation: types.TenantRelation{
			TenantID:   request.TenantID,
			Identifier: request.Identifier,
		},
		GlobalUserID: globalUserID,
	}

	domainErr := u.ketoClient.CreateRelationTuple(ctx, createRelationTupleRequest)
	if domainErr != nil {
		logger.GetLogger().Errorf("Failed to create relation tuple: %v", domainErr)
		return false, domainErr
	}
	return true, nil
}

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

	globalUserID, err := u.getGlobalUserID(ctx, &request)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get global user id: %v", err)
		return domainerrors.NewInternalError(
			"MSG_GET_GLOBAL_USER_ID_FAILED",
			"Failed to get global user id",
		)
	}

	// Set the global user id to the request
	request.GlobalUserID = globalUserID

	domainErr := u.ketoClient.CreateRelationTuple(ctx, request)
	if domainErr != nil {
		logger.GetLogger().Errorf("Failed to create relation tuple: %v", domainErr)
		return domainErr
	}

	return nil
}

func (u *permissionUseCase) getGlobalUserID(ctx context.Context, req types.PermissionRequest) (string, error) {
	identifierType, err := utils.GetIdentifierType(req.GetIdentifier())
	if err != nil {
		logger.GetLogger().Errorf("Invalid identifier: %v", err)
		return "", err
	}

	userIdentity, err := u.userIdentityRepo.GetByTypeAndValue(ctx, nil, identifierType, req.GetIdentifier())
	if err != nil {
		logger.GetLogger().Errorf("Failed to get user identity: %v", err)
		return "", domainerrors.NewInternalError(
			"MSG_GET_USER_IDENTITY_FAILED",
			"Failed to get user identity",
		)
	}
	return userIdentity.GlobalUserID, nil
}
