package interfaces

import (
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
)

type PolicyRepository interface {
	CreatePolicy(policy *domain.Policy) error
	CheckPolicyExistsByName(name string) (bool, error)
}

type PolicyUCase interface {
	CreatePolicy(payload dto.PolicyPayloadDTO) (*dto.PolicyDTO, error)
}
