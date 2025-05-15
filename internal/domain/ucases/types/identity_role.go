package interfaces

import (
	"context"

	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
)

type IdentityRoleUseCase interface {
	List(
		ctx context.Context,
		page int,
		size int,
		keyword string,
	) (*dto.PaginationDTOResponse, *dto.ErrorDTOResponse)

	GetByID(
		ctx context.Context,
		id string,
	) (*dto.IdentityRoleDTO, *dto.ErrorDTOResponse)

	Create(
		ctx context.Context,
		payload dto.CreateIdentityRolePayloadDTO,
	) (*dto.IdentityRoleDTO, *dto.ErrorDTOResponse)

	Update(
		ctx context.Context,
		id string,
		payload dto.UpdateIdentityRolePayloadDTO,
	) (*dto.IdentityRoleDTO, *dto.ErrorDTOResponse)

	Delete(
		ctx context.Context,
		id string,
	) (*dto.IdentityRoleDTO, *dto.ErrorDTOResponse)
}
