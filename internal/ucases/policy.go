package ucases

import (
	"errors"

	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
)

type PolicyUseCase struct {
	policyRepository interfaces.PolicyRepository
}

func NewPolicyUCase(policyRepo interfaces.PolicyRepository) interfaces.PolicyUCase {
	return &PolicyUseCase{policyRepository: policyRepo}
}

func (u *PolicyUseCase) CreatePolicy(payload dto.PolicyPayloadDTO) (*dto.PolicyDTO, error) {
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
