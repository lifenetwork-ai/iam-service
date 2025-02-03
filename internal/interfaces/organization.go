package interfaces

import (
	"github.com/genefriendway/human-network-iam/internal/dto"
)

type OrganizationUseCase interface {
	GetOrganizations() ([]dto.OrganizationDTO, error)
	GetOrganizationByID(organizationID string) (*dto.OrganizationDTO, error)
	CreateOrganization(organization *dto.OrganizationDTO) error
	UpdateOrganization(organization *dto.OrganizationDTO) error
	DeleteOrganization(organizationID string) error
	GetMembers(organizationID string) ([]dto.AccountDTO, error)
	AddMember(organizationID string, memberID string) error
	RemoveMember(organizationID string, memberID string) error
}

type OrganizationRepository interface {
	
}
