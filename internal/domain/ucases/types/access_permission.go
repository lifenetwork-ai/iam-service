package interfaces

import (
	"context"

	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
)

type AccessPermissionUseCase interface {
	List(
		ctx context.Context,
		page int,
		size int,
		keyword string,
	) (*dto.PaginationDTOResponse, *dto.ErrorDTOResponse)

	GetByID(
		ctx context.Context,
		id string,
	) (*dto.AccessPermissionDTO, *dto.ErrorDTOResponse)

	Create(
		ctx context.Context,
		payload dto.CreateAccessPermissionPayloadDTO,
	) (*dto.AccessPermissionDTO, *dto.ErrorDTOResponse)

	Update(
		ctx context.Context,
		id string,
		payload dto.UpdateAccessPermissionPayloadDTO,
	) (*dto.AccessPermissionDTO, *dto.ErrorDTOResponse)

	Delete(
		ctx context.Context,
		id string,
	) (*dto.AccessPermissionDTO, *dto.ErrorDTOResponse)
}
