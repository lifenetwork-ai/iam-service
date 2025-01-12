package interfaces

import (
	"github.com/genefriendway/human-network-auth/internal/domain"
)

type PolicyRepository interface {
	PolicyExists(policyID string) (bool, error)
	CreatePolicy(policy *domain.Policy) error
	CheckPolicyExistsByName(name string) (bool, error)
	GetPolicyByName(name string) (*domain.Policy, error)
}
