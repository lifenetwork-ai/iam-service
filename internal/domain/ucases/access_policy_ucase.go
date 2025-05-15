package ucases

import (
	"context"

	repositories "github.com/genefriendway/human-network-iam/internal/adapters/repositories/types"
	dto "github.com/genefriendway/human-network-iam/internal/delivery/dto"
	interfaces "github.com/genefriendway/human-network-iam/internal/domain/ucases/types"
)

type policyUseCase struct {
	policyRepo repositories.AccessPolicyRepository
}

func NewAccessPolicyUseCase(
	policyRepo repositories.AccessPolicyRepository,
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
