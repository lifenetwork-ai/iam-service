package ucases

import (
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
)

type iamUCase struct {
	iamRepository     interfaces.IAMRepository
	policyRepository  interfaces.PolicyRepository
	accountRepository interfaces.AccountRepository
}

func NewIAMUCase(
	iamRepository interfaces.IAMRepository,
	policyRepository interfaces.PolicyRepository,
	accountRepository interfaces.AccountRepository,
) interfaces.IAMUCase {
	return &iamUCase{
		iamRepository:     iamRepository,
		policyRepository:  policyRepository,
		accountRepository: accountRepository,
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
