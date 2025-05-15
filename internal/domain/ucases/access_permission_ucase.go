package ucases

import (
	"context"

	repositories "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
)

type permissionUseCase struct {
	permissionRepo repositories.AccessPermissionRepository
}

func NewAccessPermissionUseCase(
	permissionRepo repositories.AccessPermissionRepository,
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
