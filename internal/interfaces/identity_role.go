package interfaces

import (
	"context"

	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/dto"
)

type IdentityRoleUseCase interface {
	List(
		ctx context.Context,
		page int,
		size int,
		keyword string,
	) (*dto.PaginationDTOResponse, *dto.ErrorDTOResponse)

	GetByID(
		ctx context.Context,
		id string,
	) (*dto.IdentityRoleDTO, *dto.ErrorDTOResponse)

	Create(
		ctx context.Context,
		payload dto.CreateIdentityRolePayloadDTO,
	) (*dto.IdentityRoleDTO, *dto.ErrorDTOResponse)

	Update(
		ctx context.Context,
		id string,
		payload dto.UpdateIdentityRolePayloadDTO,
	) (*dto.IdentityRoleDTO, *dto.ErrorDTOResponse)

	Delete(
		ctx context.Context,
		id string,
	) (*dto.IdentityRoleDTO, *dto.ErrorDTOResponse)
}

type IdentityRoleRepository interface {
	Get(
		ctx context.Context,
		limit int,
		offset int,
		keyword *string,
	) ([]domain.IdentityRole, error)

	GetByID(
		ctx context.Context,
		id string,
	) (*domain.IdentityRole, error)

	GetByCode(
		ctx context.Context,
		code string,
	) (*domain.IdentityRole, error)

	Create(
		ctx context.Context,
		entity domain.IdentityRole,
	) (*domain.IdentityRole, error)

	Update(
		ctx context.Context,
		entity domain.IdentityRole,
	) (*domain.IdentityRole, error)

	SoftDelete(
		ctx context.Context,
		id string,
	) (*domain.IdentityRole, error)

	Delete(
		ctx context.Context,
		id string,
	) (*domain.IdentityRole, error)
}
