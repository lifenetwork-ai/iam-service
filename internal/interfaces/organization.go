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
	) ([]dto.OrganizationDTO, error)

	GetOrganizationByID(
		ctx context.Context,
		organizationID string,
	) (*dto.OrganizationDTO, error)

	CreateOrganization(
		ctx context.Context,
		payloads []dto.OrganizationCreatePayloadDTO,
	) (*dto.OrganizationDTO, error)

	UpdateOrganization(
		ctx context.Context,
		payloads []dto.OrganizationCreatePayloadDTO,
	) (*dto.OrganizationDTO, error)

	DeleteOrganization(
		ctx context.Context,
		organizationID string,
	) (*dto.OrganizationDTO, error)
}

type OrganizationRepository interface {
	GetOrganizations(
		ctx context.Context,
		limit int,
		offset int,
		keyword *string,
	) ([]domain.Organization, error)
}
