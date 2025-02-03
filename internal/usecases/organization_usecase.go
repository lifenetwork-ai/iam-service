package usecases

import (
	"github.com/genefriendway/human-network-iam/internal/dto"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type organizationUseCase struct {
	organizationRepository        interfaces.OrganizationRepository
}

func NewOrganizationUseCase(
	organizationRepository interfaces.OrganizationRepository,
) interfaces.OrganizationUseCase {
	return &organizationUseCase{
		organizationRepository:        organizationRepository,
	}
}

func (u *organizationUseCase) GetOrganizations() ([]dto.OrganizationDTO, error) {
	return nil, nil
}

func (u *organizationUseCase) GetOrganizationByID(organizationID string) (*dto.OrganizationDTO, error) {
	return nil, nil
}

func (u *organizationUseCase) CreateOrganization(organizationDTO *dto.OrganizationDTO) error {
	return nil
}

func (u *organizationUseCase) UpdateOrganization(organizationDTO *dto.OrganizationDTO) error {
	return nil
}

func (u *organizationUseCase) DeleteOrganization(organizationID string) error {
	return nil
}

func (u *organizationUseCase) GetMembers(organizationID string) ([]dto.AccountDTO, error) {
	return nil, nil
}

func (u *organizationUseCase) AddMember(organizationID string, memberID string) error {
	return nil
}

func (u *organizationUseCase) RemoveMember(organizationID string, memberID string) error {
	return nil
}

