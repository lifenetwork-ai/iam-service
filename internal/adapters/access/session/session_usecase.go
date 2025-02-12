package access_session

import (
	"context"

	"github.com/genefriendway/human-network-iam/internal/dto"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type sessionUseCase struct {
	sessionRepo interfaces.AccessSessionRepository
}

func NewAccessSessionUseCase(
	sessionRepo interfaces.AccessSessionRepository,
) interfaces.AccessSessionUseCase {
	return &sessionUseCase{
		sessionRepo: sessionRepo,
	}
}

func (u *sessionUseCase) List(
	ctx context.Context,
	page int,
	size int,
	keyword string,
) (*dto.PaginationDTOResponse, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *sessionUseCase) GetByID(
	ctx context.Context,
	id string,
) (*dto.AccessSessionDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *sessionUseCase) Create(
	ctx context.Context,
	payload dto.CreateAccessSessionPayloadDTO,
) (*dto.AccessSessionDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *sessionUseCase) Update(
	ctx context.Context,
	id string,
	payloads dto.UpdateAccessSessionPayloadDTO,
) (*dto.AccessSessionDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *sessionUseCase) Delete(
	ctx context.Context,
	id string,
) (*dto.AccessSessionDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}