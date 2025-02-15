package interfaces

import (
	"context"

	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/dto"
)

type IdentityServiceUseCase interface {
	List(
		ctx context.Context,
		page int,
		size int,
		keyword string,
	) (*dto.PaginationDTOResponse, *dto.ErrorDTOResponse)

	GetByID(
		ctx context.Context,
		id string,
	) (*dto.IdentityServiceDTO, *dto.ErrorDTOResponse)

	Create(
		ctx context.Context,
		payload dto.CreateIdentityServicePayloadDTO,
	) (*dto.IdentityServiceDTO, *dto.ErrorDTOResponse)

	Update(
		ctx context.Context,
		id string,
		payload dto.UpdateIdentityServicePayloadDTO,
	) (*dto.IdentityServiceDTO, *dto.ErrorDTOResponse)

	Delete(
		ctx context.Context,
		id string,
	) (*dto.IdentityServiceDTO, *dto.ErrorDTOResponse)
}

type IdentityServiceRepository interface {
	Get(
		ctx context.Context,
		limit int,
		offset int,
		keyword string,
	) ([]domain.IdentityService, error)

	GetByID(
		ctx context.Context,
		id string,
	) (*domain.IdentityService, error)

	GetByCode(
		ctx context.Context,
		code string,
	) (*domain.IdentityService, error)

	Create(
		ctx context.Context,
		entity domain.IdentityService,
	) (*domain.IdentityService, error)

	Update(
		ctx context.Context,
		entity domain.IdentityService,
	) (*domain.IdentityService, error)

	SoftDelete(
		ctx context.Context,
		id string,
	) (*domain.IdentityService, error)

	Delete(
		ctx context.Context,
		id string,
	) (*domain.IdentityService, error)
}
