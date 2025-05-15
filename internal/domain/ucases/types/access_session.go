package interfaces

import (
	"context"

	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
)

type AccessSessionUseCase interface {
	List(
		ctx context.Context,
		page int,
		size int,
		keyword string,
	) (*dto.PaginationDTOResponse, *dto.ErrorDTOResponse)

	GetByID(
		ctx context.Context,
		id string,
	) (*dto.AccessSessionDTO, *dto.ErrorDTOResponse)

	Delete(
		ctx context.Context,
		id string,
	) (*dto.AccessSessionDTO, *dto.ErrorDTOResponse)
}
