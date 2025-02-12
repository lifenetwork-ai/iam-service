package interfaces

import (
	"context"

	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/dto"
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

type AccessPermissionRepository interface {
	Get(
		ctx context.Context,
		limit int,
		offset int,
		keyword *string,
	) ([]domain.AccessPermission, error)

	GetByID(
		ctx context.Context,
		id string,
	) (*domain.AccessPermission, error)

	GetByCode(
		ctx context.Context,
		code string,
	) (*domain.AccessPermission, error)

	Create(
		ctx context.Context,
		entity domain.AccessPermission,
	) (*domain.AccessPermission, error)

	Update(
		ctx context.Context,
		entity domain.AccessPermission,
	) (*domain.AccessPermission, error)

	SoftDelete(
		ctx context.Context,
		id string,
	) (*domain.AccessPermission, error)

	Delete(
		ctx context.Context,
		id string,
	) (*domain.AccessPermission, error)
}
