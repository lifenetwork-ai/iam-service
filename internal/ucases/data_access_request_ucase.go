package ucases

import (
	"errors"
	"fmt"

	"github.com/genefriendway/human-network-auth/conf"
	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
	"github.com/genefriendway/human-network-auth/pkg/crypto"
)

type dataAccessUCase struct {
	config               *conf.Configuration
	dataAccessRepository interfaces.DataAccessRepository
	accountRepository    interfaces.AccountRepository
}

func NewDataAccessUCase(
	config *conf.Configuration,
	dataAccessRepository interfaces.DataAccessRepository,
	accountRepository interfaces.AccountRepository,
) interfaces.DataAccessUCase {
	return &dataAccessUCase{
		config:               config,
		dataAccessRepository: dataAccessRepository,
		accountRepository:    accountRepository,
	}
}

// CreateRequest handles the logic to create a new data access request
func (u *dataAccessUCase) CreateRequest(
	payload dto.DataAccessRequestPayloadDTO, requesterAccounts []dto.AccountDTO,
) error {
	// Ensure the requested account exists
	requestAccount, err := u.accountRepository.FindAccountByID(payload.RequestAccountID)
	if err != nil {
		return err
	}
	if requestAccount == nil {
		return errors.New("requested account not found")
	}

	// Create the DataAccessRequest domain model
	dataAccessRequest := &domain.DataAccessRequest{
		RequestAccountID: payload.RequestAccountID,
		ReasonForRequest: payload.ReasonForRequest,
		FileID:           payload.FileID,
		Status:           constants.DataAccessRequestPending.String(), // Default status
	}

	// Save the data access request in the database
	if err := u.dataAccessRepository.CreateDataAccessRequest(dataAccessRequest); err != nil {
		return err
	}

	// Loop through each requester and associate them with the request
	for _, requester := range requesterAccounts {
		requesterEntry := &domain.DataAccessRequestRequester{
			RequestID:          dataAccessRequest.ID,
			RequesterAccountID: requester.ID,
		}

		if err := u.dataAccessRepository.CreateDataAccessRequestRequester(requesterEntry); err != nil {
			return err
		}
	}

	return nil
}

// GetRequestsByStatus fetches data access requests by status
func (u *dataAccessUCase) GetRequestsByStatus(
	requestAccountID string, status constants.DataAccessRequestStatus,
) ([]dto.DataAccessRequestDTO, error) {
	// Fetch requests by status from the repository
	requests, err := u.dataAccessRepository.GetRequestsByStatus(requestAccountID, string(status))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch requests with status %s: %w", status, err)
	}

	// Process requests and convert them to DTOs
	requestDTOs := make([]dto.DataAccessRequestDTO, len(requests))
	for i, req := range requests {
		dto, err := u.populateRequesterPublicKey(req)
		if err != nil {
			return nil, err
		}
		requestDTOs[i] = *dto
	}

	return requestDTOs, nil
}

// populateRequesterPublicKey enriches a request with the public keys of all requesters
func (u *dataAccessUCase) populateRequesterPublicKey(request domain.DataAccessRequest) (*dto.DataAccessRequestDTO, error) {
	mnemonic := u.config.Secret.Mnemonic
	passphrase := u.config.Secret.Passphrase
	salt := u.config.Secret.Salt

	dto := request.ToDTO()

	// Iterate over all requesters and generate public keys
	for i, requester := range dto.Requesters {
		// Generate public key for each requester
		publicKey, _, err := crypto.GenerateAccount(mnemonic, passphrase, salt, requester.Role, requester.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate public key for requester %s: %w", requester.ID, err)
		}

		// Convert public key to hexadecimal string
		publicKeyHex, err := crypto.PublicKeyToHex(publicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to convert public key to hex for requester %s: %w", requester.ID, err)
		}

		// Assign the public key to the requester's DTO
		dto.Requesters[i].PublicKey = &publicKeyHex
	}

	return dto, nil
}

// ApproveOrRejectRequest updates the status of a data access request
func (u *dataAccessUCase) ApproveOrRejectRequestByID(
	requestAccountID, requestID string, status constants.DataAccessRequestStatus, reasonForRejection *string,
) error {
	// Validate the status
	if status != constants.DataAccessRequestApproved && status != constants.DataAccessRequestRejected {
		return errors.New("invalid status: only APPROVED or REJECTED are allowed")
	}

	// Update the request status in the repository
	if err := u.dataAccessRepository.UpdateRequestStatusByID(
		requestAccountID, requestID, status, reasonForRejection,
	); err != nil {
		return fmt.Errorf("failed to update request status: %w", err)
	}

	return nil
}

func (u *dataAccessUCase) GetRequestsByRequesterAccountID(
	requesterAccountID, status string,
) ([]dto.DataAccessRequestDTO, error) {
	// Fetch requests by requesterAccountID and status from the repository
	requests, err := u.dataAccessRepository.GetRequestsByRequesterAccountID(requesterAccountID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch requests for requester account ID %s: %w", requesterAccountID, err)
	}

	// Process requests and convert them to DTOs
	requestDTOs := make([]dto.DataAccessRequestDTO, len(requests))
	for i, req := range requests {
		dto, err := u.populateRequesterPublicKey(req)
		if err != nil {
			return nil, err
		}
		requestDTOs[i] = *dto
		requestDTOs[i].RequesterID = requesterAccountID
	}

	return requestDTOs, nil
}

func (u *dataAccessUCase) GetRequestsByRequesterAccountIDTest(
	requesterAccountID, status string,
) ([]dto.RequesterRequestDTO, error) {
	// Fetch requests by requesterAccountID and status from the repository
	requests, err := u.dataAccessRepository.GetRequestsByRequesterAccountIDTest(requesterAccountID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch requests for requester account ID %s: %w", requesterAccountID, err)
	}

	// Process requests and convert them to DTOs
	requestDTOs := make([]dto.RequesterRequestDTO, len(requests))
	for i, req := range requests {
		dto, err := u.populateRequesterPublicKey(req.Request)
		if err != nil {
			return nil, err
		}

		requestDTOs[i].DataAccessRequestDTO = *dto
		requestDTOs[i].RequesterID = requesterAccountID
		requestDTOs[i].ValidationStatus = req.ValidationStatus
		requestDTOs[i].ValidationMessage = req.ValidationMessage
	}

	return requestDTOs, nil
}

func (u *dataAccessUCase) ValidatorValidateFileContent(
	accountID, requestID string, status constants.RequesterRequestStatus, msg string,
) error {
	return u.dataAccessRepository.ValidateFileContent(accountID, requestID, status, msg)
}
