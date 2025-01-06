package interfaces

import (
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
)

type DataAccessRequestRepository interface {
	CreateDataAccessRequest(request *domain.DataAccessRequest) error
}

type DataAccessRequestUCase interface {
	CreateRequest(payload dto.DataAccessRequestPayloadDTO, requesterAccountID, requesterAccountRole string) error
}
