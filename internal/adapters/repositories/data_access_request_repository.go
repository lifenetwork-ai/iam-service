package repositories

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
)

type dataAccessRepository struct {
	db *gorm.DB
}

func NewDataAccessRepository(db *gorm.DB) interfaces.DataAccessRepository {
	return &dataAccessRepository{db: db}
}

// CreateDataAccessRequest creates a new data access request in the database.
func (r *dataAccessRepository) CreateDataAccessRequest(request *domain.DataAccessRequest) error {
	// Use GORM's Create method to insert a new data access request
	if err := r.db.Create(request).Error; err != nil {
		return err
	}
	return nil
}

// GetRequestsByStatus retrieves data access requests by requestAccountID, optionally filtered by status.
func (r *dataAccessRepository) GetRequestsByStatus(requestAccountID, status string) ([]domain.DataAccessRequest, error) {
	var requests []domain.DataAccessRequest

	// Build the query
	query := r.db.
		Preload("FileInfo.Owner").
		Preload("Requesters.RequesterAccount").
		Where("request_account_id = ?", requestAccountID)

	// Add the status condition if it is provided
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Execute the query
	err := query.Find(&requests).Error
	if err != nil {
		return nil, err
	}

	return requests, nil
}

// UpdateRequestStatus updates the status of a data access request.
// If the status is REJECTED, the reason for rejection can also be set.
func (r *dataAccessRepository) UpdateRequestStatusByID(
	requestAccountID, requestID string, status constants.DataAccessRequestStatus, reasonForRejection *string,
) error {
	// Prepare the update fields
	updateData := map[string]interface{}{
		"status": status,
	}
	if status == constants.DataAccessRequestRejected && reasonForRejection != nil {
		updateData["reason_for_rejection"] = *reasonForRejection
	}

	// Update the database record with validation by requestAccountID and requestID
	if err := r.db.Model(&domain.DataAccessRequest{}).
		Where("request_account_id = ? AND id = ?", requestAccountID, requestID).
		Updates(updateData).Error; err != nil {
		return fmt.Errorf("failed to update request status: %w", err)
	}

	return nil
}

// CreateDataAccessRequestRequester inserts a new requester entry into the data_access_request_requesters table.
func (r *dataAccessRepository) CreateDataAccessRequestRequester(requester *domain.DataAccessRequestRequester) error {
	if err := r.db.Create(requester).Error; err != nil {
		return err
	}
	return nil
}

// GetRequestsByRequesterAccountID fetches data access requests by requester id, optionally filtered by status.
func (r *dataAccessRepository) GetRequestsByRequesterAccountID(requesterAccountID, status string) ([]domain.DataAccessRequest, error) {
	var requests []domain.DataAccessRequest

	// Build the query to fetch requests by requester_account_id
	query := r.db.
		Joins("JOIN data_access_request_requesters ON data_access_request_requesters.request_id = data_access_requests.id").
		Where("data_access_request_requesters.requester_account_id = ?", requesterAccountID).
		Preload("FileInfo.Owner")

	// Add the status condition if it is provided
	if status != "" {
		query = query.Where("data_access_requests.status = ?", status)
	}

	// Execute the query
	err := query.Find(&requests).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch requests for requester account ID %s: %w", requesterAccountID, err)
	}

	return requests, nil
}

func (r *dataAccessRepository) GetRequestsByRequesterAccountIDTest(requesterAccountID, status string) ([]domain.DataAccessRequestRequesterTest, error) {
	var requesters []domain.DataAccessRequestRequesterTest

	query := r.db.Table("data_access_request_requesters").
		Where("requester_account_id = ?", requesterAccountID).
		Preload("Request.FileInfo.Owner").
		Preload("RequesterAccount")

	if status != "" {
		query = query.Where("validation_status = ?", status)
	}

	err := query.Find(&requesters).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch requests for requester account ID %s: %w", requesterAccountID, err)
	}

	return requesters, nil
}
