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

// GetRequestsByStatus fetches data access requests by status
func (u *dataAccessUCase) GetRequestsByStatus(
	requestAccountID string, status constants.DataAccessRequestStatus,
) ([]dto.DataAccessRequestDTO, error) {
	// Retrieve secret values
	mnemonic := u.config.Secret.Mnemonic
	passphrase := u.config.Secret.Passphrase
	salt := u.config.Secret.Salt

	// Fetch requests by status from the repository
	requests, err := u.dataAccessRepository.GetRequestsByStatus(requestAccountID, string(status))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch requests with status %s: %w", status, err)
	}

	// Convert domain models to DTOs
	requestDTOs := make([]dto.DataAccessRequestDTO, len(requests))
	for i, req := range requests {
		dto := req.ToDTO()

		requester := dto.RequesterAccount
		// Generate public key
		publicKey, _, err := crypto.GenerateAccount(
			mnemonic, passphrase, salt, requester.Role, requester.ID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to generate public key for requester %s: %w", requester.ID, err)
		}

		// Convert public key to hexadecimal string
		publicKeyHex, err := crypto.PublicKeyToHex(publicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to convert public key to hex for requester %s: %w", requester.ID, err)
		}
		dto.RequesterAccount.PublicKey = &publicKeyHex

		requestDTOs[i] = *dto
	}

	return requestDTOs, nil
}

// ApproveOrRejectRequest updates the status of a data access request
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

// GetAccessRequest fetches a single data access request by requestAccountID and requesterAccountID
func (u *dataAccessUCase) GetAccessRequest(requestAccountID, requesterAccountID string) (*dto.DataAccessRequestDTO, error) {
	// Fetch the approved or pending request from the repository
	request, err := u.dataAccessRepository.GetAccessRequest(requestAccountID, requesterAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch access request: %w", err)
	}

	// Handle no request found
	if request == nil {
		return nil, nil
	}

	// Convert the domain model to a DTO
	requestDTO := request.ToDTO()

	// Optionally, include generated public key for the requester
	mnemonic := u.config.Secret.Mnemonic
	passphrase := u.config.Secret.Passphrase
	salt := u.config.Secret.Salt

	publicKey, _, err := crypto.GenerateAccount(
		mnemonic, passphrase, salt, request.RequesterAccount.Role, request.RequesterAccount.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate public key for requester: %w", err)
	}

	publicKeyHex, err := crypto.PublicKeyToHex(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert public key to hex: %w", err)
	}

	requestDTO.RequesterAccount.PublicKey = &publicKeyHex

	return requestDTO, nil
}
