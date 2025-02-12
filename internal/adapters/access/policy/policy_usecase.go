package access_policy

import (
	"context"

	"github.com/genefriendway/human-network-iam/internal/dto"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type policyUseCase struct {
	policyRepo interfaces.AccessPolicyRepository
}

func NewAccessPolicyUseCase(
	policyRepo interfaces.AccessPolicyRepository,
) interfaces.AccessPolicyUseCase {
	return &policyUseCase{
		policyRepo: policyRepo,
	}
}

func (u *policyUseCase) List(
	ctx context.Context,
	page int,
	size int,
	keyword string,
) (*dto.PaginationDTOResponse, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *policyUseCase) GetByID(
	ctx context.Context,
	id string,
) (*dto.AccessPolicyDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *policyUseCase) Create(
	ctx context.Context,
	payload dto.CreateAccessPolicyPayloadDTO,
) (*dto.AccessPolicyDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *policyUseCase) Update(
	ctx context.Context,
	id string,
	payloads dto.UpdateAccessPolicyPayloadDTO,
) (*dto.AccessPolicyDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *policyUseCase) Delete(
	ctx context.Context,
	id string,
) (*dto.AccessPolicyDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}