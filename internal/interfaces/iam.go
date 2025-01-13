package interfaces

import (
	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
)

type IAMUCase interface {
	CreatePolicy(payload dto.PolicyPayloadDTO) (*dto.PolicyDTO, error)
	AssignPolicyToAccount(accountID, policyID string) error
	CheckPermission(accountID, resource, action string) (bool, error)
	CreatePermission(payload dto.PermissionPayloadDTO) error
	GetPoliciesWithPermissions() ([]dto.PolicyWithPermissionsDTO, error)
	GetPolicyByRole(role constants.AccountRole) (*dto.PolicyDTO, error)
	RemovePoliciesFromAccount(accountID string) error
}

type IAMRepository interface {
	AssignPolicyToAccount(accountID, policyID string) error
	GetAccountPermissions(accountID string) ([]domain.Permission, error)
	GetPermissionsByPolicyID(policyID string) ([]domain.Permission, error)
	RemoveAccountPolicies(accountID string) error
}
