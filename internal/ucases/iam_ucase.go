package ucases

import (
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
)

type iamUCase struct {
	iamRepository        interfaces.IAMRepository
	policyRepository     interfaces.PolicyRepository
	accountRepository    interfaces.AccountRepository
	permissionRepository interfaces.PermissionRepository
}

func NewIAMUCase(
	iamRepository interfaces.IAMRepository,
	policyRepository interfaces.PolicyRepository,
	accountRepository interfaces.AccountRepository,
	permissionRepository interfaces.PermissionRepository,
) interfaces.IAMUCase {
	return &iamUCase{
		iamRepository:        iamRepository,
		policyRepository:     policyRepository,
		accountRepository:    accountRepository,
		permissionRepository: permissionRepository,
	}
}

func (u *iamUCase) CreatePolicy(payload dto.PolicyPayloadDTO) (*dto.PolicyDTO, error) {
	// Check if the policy already exists
	exists, err := u.policyRepository.CheckPolicyExistsByName(payload.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrAlreadyExists
	}

	// Create the domain model
	policy := &domain.Policy{
		Name:        payload.Name,
		Description: payload.Description,
	}

	// Save the policy in the repository
	if err := u.policyRepository.CreatePolicy(policy); err != nil {
		return nil, err
	}

	return policy.ToDTO(), nil
}

// AssignPolicyToAccount assigns a policy to an account.
func (u *iamUCase) AssignPolicyToAccount(accountID, policyID string) error {
	accountExists, err := u.accountRepository.AccountExists(accountID)
	if err != nil || !accountExists {
		return domain.ErrDataNotFound
	}

	policyExists, err := u.policyRepository.PolicyExists(policyID)
	if err != nil || !policyExists {
		return domain.ErrDataNotFound
	}

	return u.iamRepository.AssignPolicyToAccount(accountID, policyID)
}

// CheckPermission checks if an account has permission to perform an action on a resource.
func (u *iamUCase) CheckPermission(accountID, resource, action string) (bool, error) {
	permissions, err := u.iamRepository.GetAccountPermissions(accountID)
	if err != nil {
		return false, err
	}

	for _, permission := range permissions {
		if permission.Resource == resource && permission.Action == action {
			return true, nil
		}
	}

	return false, nil
}

// CreatePermission creates a new permission for a policy.
func (u *iamUCase) CreatePermission(payload dto.PermissionPayloadDTO) error {
	if payload.PolicyID != "" && payload.PolicyName != "" {
		return domain.ErrInvalidParameters
	}

	// Fetch the policy by ID or Name
	var policy *domain.Policy
	var err error
	if payload.PolicyID != "" {
		policy, err = u.policyRepository.GetPolicyByID(payload.PolicyID)
	} else {
		policy, err = u.policyRepository.GetPolicyByName(payload.PolicyName)
	}
	if err != nil {
		if err == domain.ErrDataNotFound {
			return domain.ErrDataNotFound // Policy not found
		}
		return err // Other errors
	}

	// Check if the permission already exists
	exists, err := u.permissionRepository.PermissionExists(policy.ID, payload.Resource, payload.Action)
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrAlreadyExists
	}

	// Create the domain model for the permission
	permission := &domain.Permission{
		PolicyID:    policy.ID,
		Resource:    payload.Resource,
		Action:      payload.Action,
		Description: payload.Description,
	}

	// Save the permission in the repository
	if err := u.permissionRepository.CreatePermission(permission); err != nil {
		return err
	}

	return nil
}

// GetPoliciesWithPermissions retrieves all policies and their associated permissions.
func (u *iamUCase) GetPoliciesWithPermissions() ([]dto.PolicyWithPermissionsDTO, error) {
	// Fetch all policies
	policies, err := u.policyRepository.GetAllPolicies()
	if err != nil {
		return nil, err
	}

	// Map policies to DTOs with permissions
	var result []dto.PolicyWithPermissionsDTO
	for _, policy := range policies {
		// Fetch permissions for the policy
		permissions, err := u.iamRepository.GetPermissionsByPolicyID(policy.ID)
		if err != nil {
			return nil, err
		}

		// Convert permissions to DTOs
		permissionDTOs := make([]dto.PermissionDTO, len(permissions))
		for i, permission := range permissions {
			permissionDTOs[i] = permission.ToDTO()
		}

		// Add policy with its permissions to the result
		result = append(result, dto.PolicyWithPermissionsDTO{
			Policy:      *policy.ToDTO(),
			Permissions: permissionDTOs,
		})
	}

	return result, nil
}
