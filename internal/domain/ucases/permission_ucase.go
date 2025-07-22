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

func NewPermissionUseCase(ketoClient domainservice.KetoService, userIdentityRepo domainrepo.UserIdentityRepository) interfaces.PermissionUseCase {
	return &permissionUseCase{
		ketoClient:       ketoClient,
		userIdentityRepo: userIdentityRepo,
	}
}

// CheckPermission checks if a subject has permission to perform an action on an object.
// This function uses Ory Keto's relationship-based permission model to check access.
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

	// First check direct permission
	allowed, ketoErr := u.ketoClient.CheckPermission(ctx, request)
	if ketoErr != nil {
		logger.GetLogger().Errorf("Failed to check direct permission: %v", ketoErr)
		return false, domainerrors.NewInternalError(
			"MSG_CHECK_PERMISSION_FAILED",
			"Failed to check permission",
		)
	}

	if allowed {
		return true, nil
	}

	// If not directly allowed, check through role-based permissions
	roleRequest := types.CheckPermissionRequest{
		Namespace: "roles",
		Object:    fmt.Sprintf("%s:%s", request.TenantRelation.TenantID, u.getHigherRole(request.Relation)),
		Relation:  "member",
		TenantRelation: types.TenantRelation{
			TenantID:   request.TenantRelation.TenantID,
			Identifier: request.TenantRelation.Identifier,
		},
		GlobalUserID: globalUserID,
	}

	allowed, ketoErr = u.ketoClient.CheckPermission(ctx, roleRequest)
	if ketoErr != nil {
		logger.GetLogger().Errorf("Failed to check role permission: %v", ketoErr)
		return false, domainerrors.NewInternalError(
			"MSG_CHECK_ROLE_PERMISSION_FAILED",
			"Failed to check role permission",
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

	// Validate that the target user exists
	targetGlobalUserID, err := u.getGlobalUserIDByIdentifier(ctx, request.Identifier)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get target user global id: %v", err)
		return false, domainerrors.NewValidationError(
			"MSG_INVALID_TARGET_USER",
			"Target user not found",
			err,
		)
	}

	// Create the main permission relationship
	createRelationTupleRequest := types.CreateRelationTupleRequest{
		Namespace: request.ResourceType,
		Relation:  request.Permission,
		Object:    fmt.Sprintf("%s:%s", request.ResourceType, request.ResourceID),
		TenantRelation: types.TenantRelation{
			TenantID:   request.TenantID,
			Identifier: request.Identifier,
		},
		GlobalUserID: targetGlobalUserID,
	}

	// Also create role-based relationship if applicable
	if role := u.getRoleForPermission(request.Permission); role != "" {
		roleRequest := types.CreateRelationTupleRequest{
			Namespace: "roles",
			Relation:  "member",
			Object:    fmt.Sprintf("%s:%s", request.TenantID, role),
			TenantRelation: types.TenantRelation{
				TenantID:   request.TenantID,
				Identifier: request.Identifier,
			},
			GlobalUserID: targetGlobalUserID,
		}

		if err := u.ketoClient.CreateRelationTuple(ctx, roleRequest); err != nil {
			logger.GetLogger().Errorf("Failed to create role relation tuple: %v", err)
			return false, err
		}
	}

	if err := u.ketoClient.CreateRelationTuple(ctx, createRelationTupleRequest); err != nil {
		logger.GetLogger().Errorf("Failed to create permission relation tuple: %v", err)
		return false, err
	}

	return true, nil
}

// CreateRelationTuple creates a relation tuple for a tenant member
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

	// Create the relation tuple
	if err := u.ketoClient.CreateRelationTuple(ctx, request); err != nil {
		logger.GetLogger().Errorf("Failed to create relation tuple: %v", err)
		return err
	}

	return nil
}

func (u *permissionUseCase) getGlobalUserID(ctx context.Context, req types.PermissionRequest) (string, error) {
	identifierType, err := utils.GetIdentifierType(req.GetIdentifier())
	fmt.Println("identifierType", identifierType)
	fmt.Println("req.GetIdentifier()", req.GetIdentifier())
	if err != nil {
		logger.GetLogger().Errorf("Invalid identifier: %v", err)
		return "", err
	}
	fmt.Println("identifierType", identifierType)
	fmt.Println("req.GetIdentifier()", req.GetIdentifier())
	userIdentity, err := u.userIdentityRepo.GetByTypeAndValue(ctx, nil, identifierType, req.GetIdentifier())
	if err != nil {
		logger.GetLogger().Errorf("Failed to get user identity: %v", err)
		return "", err
	}
	return userIdentity.GlobalUserID, nil
}

func (u *permissionUseCase) getGlobalUserIDByIdentifier(ctx context.Context, identifier string) (string, error) {
	identifierType, err := utils.GetIdentifierType(identifier)
	if err != nil {
		logger.GetLogger().Errorf("Invalid identifier: %v", err)
		return "", err
	}

	userIdentity, err := u.userIdentityRepo.GetByTypeAndValue(ctx, nil, identifierType, identifier)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get user identity: %v", err)
		return "", err
	}
	return userIdentity.GlobalUserID, nil
}

// getHigherRole returns the role that implicitly grants the requested permission
func (u *permissionUseCase) getHigherRole(permission string) string {
	switch permission {
	case "view":
		return "viewer"
	case "edit", "update":
		return "editor"
	case "delete", "manage":
		return "admin"
	default:
		return ""
	}
}

// getRoleForPermission returns the role associated with a permission
func (u *permissionUseCase) getRoleForPermission(permission string) string {
	switch permission {
	case "view":
		return "viewer"
	case "edit":
		return "editor"
	case "manage":
		return "admin"
	default:
		return ""
	}
}
