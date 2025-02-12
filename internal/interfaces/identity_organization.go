package interfaces

import (
	"context"

	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/dto"
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

type IdentityOrganizationRepository interface {
	Get(
		ctx context.Context,
		limit int,
		offset int,
		keyword *string,
	) ([]domain.IdentityOrganization, error)

	GetByID(
		ctx context.Context,
		id string,
	) (*domain.IdentityOrganization, error)

	GetByCode(
		ctx context.Context,
		code string,
	) (*domain.IdentityOrganization, error)

	Create(
		ctx context.Context,
		entity domain.IdentityOrganization,
	) (*domain.IdentityOrganization, error)

	Update(
		ctx context.Context,
		entity domain.IdentityOrganization,
	) (*domain.IdentityOrganization, error)

	SoftDelete(
		ctx context.Context,
		id string,
	) (*domain.IdentityOrganization, error)

	Delete(
		ctx context.Context,
		id string,
	) (*domain.IdentityOrganization, error)
}
