package interfaces

import (
	"context"

	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/dto"
)

type AccessPolicyUseCase interface {
	List(
		ctx context.Context,
		page int,
		size int,
		keyword string,
	) (*dto.PaginationDTOResponse, *dto.ErrorDTOResponse)

	GetByID(
		ctx context.Context,
		id string,
	) (*dto.AccessPolicyDTO, *dto.ErrorDTOResponse)

	Create(
		ctx context.Context,
		payload dto.CreateAccessPolicyPayloadDTO,
	) (*dto.AccessPolicyDTO, *dto.ErrorDTOResponse)

	Update(
		ctx context.Context,
		id string,
		payload dto.UpdateAccessPolicyPayloadDTO,
	) (*dto.AccessPolicyDTO, *dto.ErrorDTOResponse)

	Delete(
		ctx context.Context,
		id string,
	) (*dto.AccessPolicyDTO, *dto.ErrorDTOResponse)
}

type AccessPolicyRepository interface {
	Get(
		ctx context.Context,
		limit int,
		offset int,
		keyword *string,
	) ([]domain.AccessPolicy, error)

	GetByID(
		ctx context.Context,
		id string,
	) (*domain.AccessPolicy, error)

	GetByCode(
		ctx context.Context,
		code string,
	) (*domain.AccessPolicy, error)

	Create(
		ctx context.Context,
		entity domain.AccessPolicy,
	) (*domain.AccessPolicy, error)

	Update(
		ctx context.Context,
		entity domain.AccessPolicy,
	) (*domain.AccessPolicy, error)

	SoftDelete(
		ctx context.Context,
		id string,
	) (*domain.AccessPolicy, error)

	Delete(
		ctx context.Context,
		id string,
	) (*domain.AccessPolicy, error)
}
