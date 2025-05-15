package interfaces

import (
	"context"

	"github.com/genefriendway/human-network-iam/internal/delivery/dto"
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
