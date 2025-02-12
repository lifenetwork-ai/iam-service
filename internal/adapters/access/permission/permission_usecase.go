package access_permission

import (
	"context"

	"github.com/genefriendway/human-network-iam/internal/dto"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type permissionUseCase struct {
	permissionRepo interfaces.AccessPermissionRepository
}

func NewAccessPermissionUseCase(
	permissionRepo interfaces.AccessPermissionRepository,
) interfaces.AccessPermissionUseCase {
	return &permissionUseCase{
		permissionRepo: permissionRepo,
	}
}

func (u *permissionUseCase) List(
	ctx context.Context,
	page int,
	size int,
	keyword string,
) (*dto.PaginationDTOResponse, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *permissionUseCase) GetByID(
	ctx context.Context,
	id string,
) (*dto.AccessPermissionDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *permissionUseCase) Create(
	ctx context.Context,
	payload dto.CreateAccessPermissionPayloadDTO,
) (*dto.AccessPermissionDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *permissionUseCase) Update(
	ctx context.Context,
	id string,
	payloads dto.UpdateAccessPermissionPayloadDTO,
) (*dto.AccessPermissionDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *permissionUseCase) Delete(
	ctx context.Context,
	id string,
) (*dto.AccessPermissionDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}