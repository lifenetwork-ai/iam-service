package interfaces

import "github.com/genefriendway/human-network-auth/internal/dto"

type IAMUCase interface {
	CreatePolicy(payload dto.PolicyPayloadDTO) (*dto.PolicyDTO, error)
	AssignPolicyToAccount(accountID, policyID string) error
}
