package identity_organization

import (
	"context"
	"fmt"
	"net/http"

	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/dto"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
	"github.com/genefriendway/human-network-iam/packages/logger"
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
) (*dto.PaginationDTOResponse, *dto.ErrorDTOResponse) {
	// Setup pagination variables
	limit := size + 1 // Fetch one extra record to determine if there's a next page
	var offset int
	if page > 0 {
		offset = page * size
	} else {
		offset = size
	}

	organizations, err := u.organizationRepo.Get(ctx, limit, offset, &keyword)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Details: []interface{}{err},
		}
	}

	// Determine if there's a next page
	nextPage := page
	if len(organizations) > size {
		nextPage += 1
	}

	var organizationDTOs []interface{} = make([]interface{}, 0)
	for _, organization := range organizations {
		organizationDTOs = append(organizationDTOs, organization.ToDTO())
	}

	// Return the response DTO
	return &dto.PaginationDTOResponse{
		NextPage: nextPage,
		Page:     page,
		Size:     size,
		Data:     organizationDTOs,
	}, nil
}

func (u *organizationUseCase) GetOrganizationByID(
	ctx context.Context,
	organizationID string,
) (*dto.OrganizationDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *organizationUseCase) CreateOrganization(
	ctx context.Context,
	payload dto.OrganizationCreatePayloadDTO,
) (*dto.OrganizationDTO, *dto.ErrorDTOResponse) {
	// Check if the organization already exists
	exist, err := u.organizationRepo.GetByCode(ctx, payload.Code)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Details: []interface{}{err},
		}
	}

	if exist != nil {
		logger.GetLogger().Errorf("Organization with code %s already exists", payload.Code)
		return nil, &dto.ErrorDTOResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Organization with code %s already exists", payload.Code),
			Details: []interface{}{fmt.Sprintf("Organization with code %s already exists", payload.Code)},
		}
	}

	// Create the organization
	newOrganization := domain.Organization{
		Name:        payload.Name,
		Code:        payload.Code,
		Description: payload.Description,
	}

	if payload.ParentID != "" {
		// Check if the parent organization exists
		parentOrganization, err := u.organizationRepo.GetByID(ctx, payload.ParentID)
		if err != nil {
			return nil, &dto.ErrorDTOResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
				Details: []interface{}{err},
			}
		}

		if parentOrganization == nil {
			logger.GetLogger().Errorf("Parent organization with ID %s does not exist", payload.ParentID)
			return nil, &dto.ErrorDTOResponse{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("Parent organization with ID %s does not exist", payload.ParentID),
				Details: []interface{}{fmt.Sprintf("Parent organization with ID %s does not exist", payload.ParentID)},
			}
		}

		newOrganization.ParentID = payload.ParentID
		newOrganization.ParentPath = parentOrganization.ParentPath + "::" + parentOrganization.ID
	}

	organization, err := u.organizationRepo.Create(ctx, newOrganization)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Details: []interface{}{err},
		}
	}

	// Return the response DTO
	dto := organization.ToDTO()
	return &dto, nil
}

func (u *organizationUseCase) UpdateOrganization(
	ctx context.Context,
	payloads dto.OrganizationUpdatePayloadDTO,
) (*dto.OrganizationDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *organizationUseCase) DeleteOrganization(
	ctx context.Context,
	organizationID string,
) (*dto.OrganizationDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}
