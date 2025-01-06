package ucases

import (
	"errors"
	"fmt"

	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
)

type dataAccessUCase struct {
	dataAccessRepository interfaces.DataAccessRepository
	accountRepository    interfaces.AccountRepository
}

func NewDataAccessUCase(
	dataAccessRepository interfaces.DataAccessRepository,
	accountRepository interfaces.AccountRepository,
) interfaces.DataAccessUCase {
	return &dataAccessUCase{
		dataAccessRepository: dataAccessRepository,
		accountRepository:    accountRepository,
	}
}

// CreateRequest handles the logic to create a new data access request
func (u *dataAccessUCase) CreateRequest(
	payload dto.DataAccessRequestPayloadDTO, requesterAccountID, requesterAccountRole string,
) error {
	// Ensure the requester and requested accounts exist
	requestAccount, err := u.accountRepository.FindAccountByID(payload.RequestAccountID)
	if err != nil {
		return err
	}
	if requestAccount == nil {
		return errors.New("requested account not found")
	}

	// Create the DataAccessRequest domain model
	dataAccessRequest := &domain.DataAccessRequest{
		RequestAccountID:   payload.RequestAccountID,
		RequesterAccountID: requesterAccountID,
		RequesterRole:      requesterAccountRole,
		ReasonForRequest:   payload.ReasonForRequest,
		Status:             string(constants.DataAccessRequestPending), // Default status
	}

	// Save the request in the database using the repository
	if err := u.dataAccessRepository.CreateDataAccessRequest(dataAccessRequest); err != nil {
		return err
	}

	return nil
}

func (u *dataAccessUCase) GetPendingRequests(userID string) ([]dto.DataAccessRequestDTO, error) {
	// Fetch pending requests from the repository
	requests, err := u.dataAccessRepository.GetPendingRequests(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending requests: %w", err)
	}

	// Convert domain models to DTOs
	requestDTOs := make([]dto.DataAccessRequestDTO, len(requests))
	for i, req := range requests {
		requestDTOs[i] = *req.ToDTO()
	}

	return requestDTOs, nil
}

func (u *dataAccessUCase) ApproveOrRejectRequest(
	requestAccountID, requesterAccountID string, status constants.DataAccessRequestStatus, reasonForRejection *string,
) error {
	// Validate the status
	if status != constants.DataAccessRequestApproved && status != constants.DataAccessRequestRejected {
		return errors.New("invalid status: only APPROVED or REJECTED are allowed")
	}

	// Update the request status in the repository
	if err := u.dataAccessRepository.UpdateRequestStatus(
		requestAccountID, requesterAccountID, status, reasonForRejection,
	); err != nil {
		return fmt.Errorf("failed to update request status: %w", err)
	}

	return nil
}
