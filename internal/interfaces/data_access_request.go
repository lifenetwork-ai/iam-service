package interfaces

import (
	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
)

type DataAccessRepository interface {
	CreateDataAccessRequest(request *domain.DataAccessRequest) error
	GetRequestsByStatus(requestAccountID, status string) ([]domain.DataAccessRequest, error)
	UpdateRequestStatusByID(
		requestAccountID, requestID string, status constants.DataAccessRequestStatus, reasonForRejection *string,
	) error
	CreateDataAccessRequestRequester(requester *domain.DataAccessRequestRequester) error
	GetRequestsByRequesterAccountID(requesterAccountID, status string) ([]domain.DataAccessRequest, error)

	GetRequestsByRequesterAccountIDTest(requesterAccountID, status string) ([]domain.DataAccessRequestRequesterTest, error)

	// TODO: sepearte this later / naming
	ValidateFileContent(accountID, requestID string, status constants.RequesterRequestStatus, msg string) error
}

type DataAccessUCase interface {
	CreateRequest(payload dto.DataAccessRequestPayloadDTO, requesterAccounts []dto.AccountDTO) error
	GetRequestsByStatus(
		requestAccountID string, status constants.DataAccessRequestStatus,
	) ([]dto.DataAccessRequestDTO, error)
	ApproveOrRejectRequestByID(
		requestAccountID, requestID string, status constants.DataAccessRequestStatus, reasonForRejection *string,
	) error
	GetRequestsByRequesterAccountID(
		requesterAccountID, status string,
	) ([]dto.DataAccessRequestDTO, error)

	GetRequestsByRequesterAccountIDTest(
		requesterAccountID, status string,
	) ([]dto.RequesterRequestDTO, error)

	// TODO: sepearte this later / naming
	ValidatorValidateFileContent(accountID, requestID string, status constants.RequesterRequestStatus, msg string) error
}
