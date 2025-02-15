package ucases

import (
	"context"

	"github.com/genefriendway/human-network-iam/internal/dto"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type roleUseCase struct {
	roleRepo interfaces.IdentityRoleRepository
}

func NewIdentityRoleUseCase(
	roleRepo interfaces.IdentityRoleRepository,
) interfaces.IdentityRoleUseCase {
	return &roleUseCase{
		roleRepo: roleRepo,
	}
}

func (u *roleUseCase) List(
	ctx context.Context,
	page int,
	size int,
	keyword string,
) (*dto.PaginationDTOResponse, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *roleUseCase) GetByID(
	ctx context.Context,
	id string,
) (*dto.IdentityRoleDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *roleUseCase) Create(
	ctx context.Context,
	payload dto.CreateIdentityRolePayloadDTO,
) (*dto.IdentityRoleDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *roleUseCase) Update(
	ctx context.Context,
	id string,
	payloads dto.UpdateIdentityRolePayloadDTO,
) (*dto.IdentityRoleDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *roleUseCase) Delete(
	ctx context.Context,
	id string,
) (*dto.IdentityRoleDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}
