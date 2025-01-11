package ucases

import (
	"errors"

	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
)

type iamUCase struct {
	policyRepository  interfaces.PolicyRepository
	accountRepository interfaces.AccountRepository
}

func NewIAMUCase(
	policyRepository interfaces.PolicyRepository,
	accountRepository interfaces.AccountRepository,
) interfaces.IAMUCase {
	return &iamUCase{
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
		return nil, errors.New("policy with this name already exists")
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

	return u.policyRepository.AssignPolicyToAccount(accountID, policyID)
}
