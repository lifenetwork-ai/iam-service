package interfaces

import (
	"context"

	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/dto"
)

type IdentityGroupUseCase interface {
	List(
		ctx context.Context,
		page int,
		size int,
		keyword string,
	) (*dto.PaginationDTOResponse, *dto.ErrorDTOResponse)

	GetByID(
		ctx context.Context,
		id string,
	) (*dto.IdentityGroupDTO, *dto.ErrorDTOResponse)

	Create(
		ctx context.Context,
		payload dto.CreateIdentityGroupPayloadDTO,
	) (*dto.IdentityGroupDTO, *dto.ErrorDTOResponse)

	Update(
		ctx context.Context,
		id string,
		payload dto.UpdateIdentityGroupPayloadDTO,
	) (*dto.IdentityGroupDTO, *dto.ErrorDTOResponse)

	Delete(
		ctx context.Context,
		id string,
	) (*dto.IdentityGroupDTO, *dto.ErrorDTOResponse)
}

type IdentityGroupRepository interface {
	Get(
		ctx context.Context,
		limit int,
		offset int,
		keyword *string,
	) ([]domain.IdentityGroup, error)

	GetByID(
		ctx context.Context,
		id string,
	) (*domain.IdentityGroup, error)

	GetByCode(
		ctx context.Context,
		code string,
	) (*domain.IdentityGroup, error)

	Create(
		ctx context.Context,
		entity domain.IdentityGroup,
	) (*domain.IdentityGroup, error)

	Update(
		ctx context.Context,
		entity domain.IdentityGroup,
	) (*domain.IdentityGroup, error)

	SoftDelete(
		ctx context.Context,
		id string,
	) (*domain.IdentityGroup, error)

	Delete(
		ctx context.Context,
		id string,
	) (*domain.IdentityGroup, error)
}
