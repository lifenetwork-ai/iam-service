package identity_organization

import (
	"context"

	"github.com/genefriendway/human-network-iam/internal/dto"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type organizationUseCase struct {
	organizationRepo interfaces.OrganizationRepository
}

func NewOrganizationUseCase(
	organizationRepo interfaces.OrganizationRepository,
) interfaces.OrganizationUseCase {
	return &organizationUseCase{
		organizationRepo: organizationRepo,
	}
}

func (u *organizationUseCase) GetOrganizations(
	ctx context.Context,
	page int,
	size int,
	keyword string,
) ([]dto.OrganizationDTO, error) {
	organizations, error := u.organizationRepo.GetOrganizations(ctx, page, size, &keyword)
	if error != nil {
		return nil, error
	}

	organizationDTOs := make([]dto.OrganizationDTO, 0)
	for _, organization := range organizations {
		organizationDTOs = append(organizationDTOs, dto.OrganizationDTO{
			ID:          organization.ID,
			Name:        organization.Name,
			Code:        organization.Code,
			Description: organization.Description,
		})
	}

	return organizationDTOs, nil
}

func (u *organizationUseCase) GetOrganizationByID(
	ctx context.Context,
	organizationID string,
) (*dto.OrganizationDTO, error) {
	return nil, nil
}

func (u *organizationUseCase) CreateOrganization(
	ctx context.Context,
	payloads []dto.OrganizationCreatePayloadDTO,
) (*dto.OrganizationDTO, error) {
	return nil, nil
}

func (u *organizationUseCase) UpdateOrganization(
	ctx context.Context,
	payloads []dto.OrganizationCreatePayloadDTO,
) (*dto.OrganizationDTO, error) {
	return nil, nil
}

func (u *organizationUseCase) DeleteOrganization(
	ctx context.Context,
	organizationID string,
) (*dto.OrganizationDTO, error) {
	return nil, nil
}
