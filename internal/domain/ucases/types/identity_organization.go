package interfaces

import (
	"context"

	"github.com/genefriendway/human-network-iam/internal/delivery/dto"
)

type IdentityOrganizationUseCase interface {
	List(
		ctx context.Context,
		page int,
		size int,
		keyword string,
	) (*dto.PaginationDTOResponse, *dto.ErrorDTOResponse)

	GetByID(
		ctx context.Context,
		id string,
	) (*dto.IdentityOrganizationDTO, *dto.ErrorDTOResponse)

	Create(
		ctx context.Context,
		payload dto.CreateIdentityOrganizationPayloadDTO,
	) (*dto.IdentityOrganizationDTO, *dto.ErrorDTOResponse)

	Update(
		ctx context.Context,
		id string,
		payload dto.UpdateIdentityOrganizationPayloadDTO,
	) (*dto.IdentityOrganizationDTO, *dto.ErrorDTOResponse)

	Delete(
		ctx context.Context,
		id string,
	) (*dto.IdentityOrganizationDTO, *dto.ErrorDTOResponse)
}
