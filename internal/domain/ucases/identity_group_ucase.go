package ucases

import (
	"context"

	repositories "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
)

type groupUseCase struct {
	groupRepo repositories.IdentityGroupRepository
}

func NewIdentityGroupUseCase(
	groupRepo repositories.IdentityGroupRepository,
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
