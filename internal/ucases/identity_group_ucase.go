package ucases

import (
	"context"

	"github.com/genefriendway/human-network-iam/internal/dto"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type groupUseCase struct {
	groupRepo interfaces.IdentityGroupRepository
}

func NewIdentityGroupUseCase(
	groupRepo interfaces.IdentityGroupRepository,
) interfaces.IdentityGroupUseCase {
	return &groupUseCase{
		groupRepo: groupRepo,
	}
}

func (u *groupUseCase) List(
	ctx context.Context,
	page int,
	size int,
	keyword string,
) (*dto.PaginationDTOResponse, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *groupUseCase) GetByID(
	ctx context.Context,
	id string,
) (*dto.IdentityGroupDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *groupUseCase) Create(
	ctx context.Context,
	payload dto.CreateIdentityGroupPayloadDTO,
) (*dto.IdentityGroupDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *groupUseCase) Update(
	ctx context.Context,
	id string,
	payloads dto.UpdateIdentityGroupPayloadDTO,
) (*dto.IdentityGroupDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *groupUseCase) Delete(
	ctx context.Context,
	id string,
) (*dto.IdentityGroupDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}
