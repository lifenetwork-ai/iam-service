package interfaces

import (
	"context"

	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/dto"
)

type OrganizationUseCase interface {
	GetOrganizations(
		ctx context.Context,
		page int,
		size int,
		keyword string,
	) (*dto.PaginationDTOResponse, *dto.ErrorDTOResponse)

	GetOrganizationByID(
		ctx context.Context,
		organizationID string,
	) (*dto.OrganizationDTO, *dto.ErrorDTOResponse)

	CreateOrganization(
		ctx context.Context,
		payload dto.OrganizationCreatePayloadDTO,
	) (*dto.OrganizationDTO, *dto.ErrorDTOResponse)

	UpdateOrganization(
		ctx context.Context,
		payload dto.OrganizationUpdatePayloadDTO,
	) (*dto.OrganizationDTO, *dto.ErrorDTOResponse)

	DeleteOrganization(
		ctx context.Context,
		organizationID string,
	) (*dto.OrganizationDTO, *dto.ErrorDTOResponse)
}

type OrganizationRepository interface {
	Get(
		ctx context.Context,
		limit int,
		offset int,
		keyword *string,
	) ([]domain.Organization, error)

	GetByID(
		ctx context.Context,
		id string,
	) (*domain.Organization, error)

	GetByCode(
		ctx context.Context,
		code string,
	) (*domain.Organization, error)

	Create(
		ctx context.Context,
		organization domain.Organization,
	) (*domain.Organization, error)

	Update(
		ctx context.Context,
		organization domain.Organization,
	) (*domain.Organization, error)

	Delete(
		ctx context.Context,
		id string,
	) (*domain.Organization, error)
}
