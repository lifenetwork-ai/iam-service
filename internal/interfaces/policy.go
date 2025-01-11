package interfaces

import (
	"github.com/genefriendway/human-network-auth/internal/domain"
)

type PolicyRepository interface {
	CreatePolicy(policy *domain.Policy) error
	CheckPolicyExistsByName(name string) (bool, error)
}
