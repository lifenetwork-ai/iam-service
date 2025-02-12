package interfaces

import (
	"context"

	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/dto"
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

	Create(
		ctx context.Context,
		payload dto.CreateAccessSessionPayloadDTO,
	) (*dto.AccessSessionDTO, *dto.ErrorDTOResponse)

	Update(
		ctx context.Context,
		id string,
		payload dto.UpdateAccessSessionPayloadDTO,
	) (*dto.AccessSessionDTO, *dto.ErrorDTOResponse)

	Delete(
		ctx context.Context,
		id string,
	) (*dto.AccessSessionDTO, *dto.ErrorDTOResponse)
}

type AccessSessionRepository interface {
	Get(
		ctx context.Context,
		limit int,
		offset int,
		keyword *string,
	) ([]domain.AccessSession, error)

	GetByID(
		ctx context.Context,
		id string,
	) (*domain.AccessSession, error)

	GetByCode(
		ctx context.Context,
		code string,
	) (*domain.AccessSession, error)

	Create(
		ctx context.Context,
		entity domain.AccessSession,
	) (*domain.AccessSession, error)

	Update(
		ctx context.Context,
		entity domain.AccessSession,
	) (*domain.AccessSession, error)

	SoftDelete(
		ctx context.Context,
		id string,
	) (*domain.AccessSession, error)

	Delete(
		ctx context.Context,
		id string,
	) (*domain.AccessSession, error)
}
