package ucases

import (
	"context"

	repositories "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
)

type sessionUseCase struct {
	sessionRepo repositories.AccessSessionRepository
}

func NewAccessSessionUseCase(
	sessionRepo repositories.AccessSessionRepository,
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

func (u *sessionUseCase) Delete(
	ctx context.Context,
	id string,
) (*dto.AccessSessionDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}
