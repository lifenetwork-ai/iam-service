package ucases

import (
	"context"

	"github.com/genefriendway/human-network-iam/internal/dto"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type serviceUseCase struct {
	serviceRepo interfaces.IdentityServiceRepository
}

func NewIdentityServiceUseCase(
	serviceRepo interfaces.IdentityServiceRepository,
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
