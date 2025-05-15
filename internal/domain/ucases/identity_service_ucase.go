package ucases

import (
	"context"

	repositories "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
)

type serviceUseCase struct {
	serviceRepo repositories.IdentityServiceRepository
}

func NewIdentityServiceUseCase(
	serviceRepo repositories.IdentityServiceRepository,
) interfaces.IdentityServiceUseCase {
	return &serviceUseCase{
		serviceRepo: serviceRepo,
	}
}

func (u *serviceUseCase) List(
	ctx context.Context,
	page int,
	size int,
	keyword string,
) (*dto.PaginationDTOResponse, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *serviceUseCase) GetByID(
	ctx context.Context,
	id string,
) (*dto.IdentityServiceDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *serviceUseCase) Create(
	ctx context.Context,
	payload dto.CreateIdentityServicePayloadDTO,
) (*dto.IdentityServiceDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *serviceUseCase) Update(
	ctx context.Context,
	id string,
	payloads dto.UpdateIdentityServicePayloadDTO,
) (*dto.IdentityServiceDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}

func (u *serviceUseCase) Delete(
	ctx context.Context,
	id string,
) (*dto.IdentityServiceDTO, *dto.ErrorDTOResponse) {
	return nil, nil
}
