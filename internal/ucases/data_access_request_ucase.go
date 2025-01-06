package ucases

import (
	"errors"

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
