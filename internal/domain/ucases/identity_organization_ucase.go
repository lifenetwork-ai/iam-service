package ucases

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	repositories "github.com/genefriendway/human-network-iam/internal/adapters/repositories/types"
	dto "github.com/genefriendway/human-network-iam/internal/delivery/dto"
	domain "github.com/genefriendway/human-network-iam/internal/domain/entities"
	interfaces "github.com/genefriendway/human-network-iam/internal/domain/ucases/types"
	"github.com/genefriendway/human-network-iam/packages/logger"
)

type organizationUseCase struct {
	organizationRepo repositories.IdentityOrganizationRepository
}

func NewIdentityOrganizationUseCase(
	organizationRepo repositories.IdentityOrganizationRepository,
) interfaces.IdentityOrganizationUseCase {
	return &organizationUseCase{
		organizationRepo: organizationRepo,
	}
}

func (u *organizationUseCase) List(
	ctx context.Context,
	page int,
	size int,
	keyword string,
) (*dto.PaginationDTOResponse, *dto.ErrorDTOResponse) {
	// Setup pagination variables
	limit := size + 1 // Fetch one extra record to determine if there's a next page
	offset := (page - 1) * size

	organizations, err := u.organizationRepo.Get(ctx, limit, offset, keyword)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
			Details: []interface{}{err},
		}
	}

	// Determine if there's a next page
	nextPage := page
	if len(organizations) > size {
		nextPage += 1
	}

	organizationDTOs := make([]interface{}, 0)
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

func (u *organizationUseCase) GetByID(
	ctx context.Context,
	id string,
) (*dto.IdentityOrganizationDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *organizationUseCase) Create(
	ctx context.Context,
	payload dto.CreateIdentityOrganizationPayloadDTO,
) (*dto.IdentityOrganizationDTO, *dto.ErrorDTOResponse) {
	// Check if the organization already exists
	exist, err := u.organizationRepo.GetByCode(ctx, payload.Code)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
			Details: []interface{}{err},
		}
	}

	if exist != nil {
		logger.GetLogger().Errorf("Organization with code %s already exists", payload.Code)
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Organization with code %s already exists", payload.Code),
			Details: []interface{}{fmt.Sprintf("Organization with code %s already exists", payload.Code)},
		}
	}

	// Create the organization
	newOrganization := domain.IdentityOrganization{
		Name:        strings.TrimSpace(payload.Name),
		Code:        strings.ToUpper(payload.Code),
		Description: strings.TrimSpace(payload.Description),
	}

	if payload.ParentID != "" {
		// Check if the parent organization exists
		parentOrganization, err := u.organizationRepo.GetByID(ctx, payload.ParentID)
		if err != nil {
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusBadRequest,
				Message: err.Error(),
				Details: []interface{}{err},
			}
		}

		if parentOrganization == nil {
			logger.GetLogger().Errorf("Parent organization with ID %s does not exist", payload.ParentID)
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusBadRequest,
				Message: fmt.Sprintf("Parent organization with ID %s does not exist", payload.ParentID),
				Details: []interface{}{fmt.Sprintf("Parent organization with ID %s does not exist", payload.ParentID)},
			}
		}

		newOrganization.ParentID = payload.ParentID
		if parentOrganization.ParentPath == "" {
			newOrganization.ParentPath = parentOrganization.Code
		} else {
			newOrganization.ParentPath = parentOrganization.ParentPath + "::" + parentOrganization.Code
		}
	}

	organization, err := u.organizationRepo.Create(ctx, newOrganization)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
			Details: []interface{}{err},
		}
	}

	// Return the response DTO
	dto := organization.ToDTO()
	return &dto, nil
}

func (u *organizationUseCase) Update(
	ctx context.Context,
	id string,
	payloads dto.UpdateIdentityOrganizationPayloadDTO,
) (*dto.IdentityOrganizationDTO, *dto.ErrorDTOResponse) {
	exist, err := u.organizationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
			Details: []interface{}{err},
		}
	}

	if exist == nil {
		logger.GetLogger().Errorf("Organization with ID %s does not exist", id)
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Organization with ID %s does not exist", id),
			Details: []interface{}{fmt.Sprintf("Organization with ID %s does not exist", id)},
		}
	}

	hasUpdate := false
	// Update the organization
	if payloads.Name != "" && payloads.Name != exist.Name {
		exist.Name = payloads.Name
		hasUpdate = true
	}

	if payloads.Code != "" && payloads.Code != exist.Code {
		// Check if the organization already exists
		existCode, err := u.organizationRepo.GetByCode(ctx, payloads.Code)
		if err != nil {
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusBadRequest,
				Message: err.Error(),
				Details: []interface{}{err},
			}
		}

		if existCode != nil {
			logger.GetLogger().Errorf("Organization with code %s already exists", payloads.Code)
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusBadRequest,
				Message: fmt.Sprintf("Organization with code %s already exists", payloads.Code),
				Details: []interface{}{fmt.Sprintf("Organization with code %s already exists", payloads.Code)},
			}
		}

		exist.Code = payloads.Code
		hasUpdate = true
	}

	if payloads.Description != "" && payloads.Description != exist.Description {
		exist.Description = payloads.Description
		hasUpdate = true
	}

	if payloads.ParentID != "" && payloads.ParentID != exist.ParentID {
		// Check if the parent organization exists
		parentOrganization, err := u.organizationRepo.GetByID(ctx, payloads.ParentID)
		if err != nil {
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusBadRequest,
				Message: err.Error(),
				Details: []interface{}{err},
			}
		}

		if parentOrganization == nil {
			logger.GetLogger().Errorf("Parent organization with ID %s does not exist", payloads.ParentID)
			return nil, &dto.ErrorDTOResponse{
				Status:  http.StatusBadRequest,
				Message: fmt.Sprintf("Parent organization with ID %s does not exist", payloads.ParentID),
				Details: []interface{}{fmt.Sprintf("Parent organization with ID %s does not exist", payloads.ParentID)},
			}
		}

		exist.ParentID = payloads.ParentID
		if parentOrganization.ParentPath == "" {
			exist.ParentPath = parentOrganization.Code
		} else {
			exist.ParentPath = parentOrganization.ParentPath + "::" + parentOrganization.Code
		}
		hasUpdate = true
	}

	if !hasUpdate {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Message: "No update data",
			Details: []interface{}{"No update data"},
		}
	}

	organization, err := u.organizationRepo.Update(ctx, *exist)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
			Details: []interface{}{err},
		}
	}

	// Return the response DTO
	dto := organization.ToDTO()
	return &dto, nil
}

func (u *organizationUseCase) Delete(
	ctx context.Context,
	id string,
) (*dto.IdentityOrganizationDTO, *dto.ErrorDTOResponse) {
	organization, err := u.organizationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
			Details: []interface{}{err},
		}
	}

	if organization == nil {
		logger.GetLogger().Errorf("Organization with ID %s does not exist", id)
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Organization with ID %s does not exist", id),
			Details: []interface{}{fmt.Sprintf("Organization with ID %s does not exist", id)},
		}
	}

	organization, err = u.organizationRepo.SoftDelete(ctx, id)
	if err != nil {
		return nil, &dto.ErrorDTOResponse{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
			Details: []interface{}{err},
		}
	}

	// Return the response DTO
	dto := organization.ToDTO()
	return &dto, nil
}
