package interfaces

import (
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
)

type DataAccessRepository interface {
	CreateDataAccessRequest(request *domain.DataAccessRequest) error
	GetPendingRequests(userID string) ([]domain.DataAccessRequest, error)
}

type DataAccessUCase interface {
	CreateRequest(payload dto.DataAccessRequestPayloadDTO, requesterAccountID, requesterAccountRole string) error
	GetPendingRequests(userID string) ([]dto.DataAccessRequestDTO, error)
}
