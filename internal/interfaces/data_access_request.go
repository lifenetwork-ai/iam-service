package interfaces

import (
	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
)

type DataAccessRepository interface {
	CreateDataAccessRequest(request *domain.DataAccessRequest) error
	GetRequestsByStatus(requestAccountID, status string) ([]domain.DataAccessRequest, error)
	UpdateRequestStatus(
		requestAccountID, requesterAccountID string, status constants.DataAccessRequestStatus, reasonForRejection *string,
	) error
	GetAccessRequest(requestAccountID, requesterAccountID string) (*domain.DataAccessRequest, error)
}

type DataAccessUCase interface {
	CreateRequest(payload dto.DataAccessRequestPayloadDTO, requesterAccountID, requesterAccountRole string) error
	GetRequestsByStatus(
		requestAccountID string, status constants.DataAccessRequestStatus,
	) ([]dto.DataAccessRequestDTO, error)
	ApproveOrRejectRequest(
		requestAccountID, requesterAccountID string, status constants.DataAccessRequestStatus, reasonForRejection *string,
	) error
	GetAccessRequest(requestAccountID, requesterAccountID string) (*dto.DataAccessRequestDTO, error)
}
