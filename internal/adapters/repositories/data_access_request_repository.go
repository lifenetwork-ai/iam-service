package repositories

import (
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

// GetRequestsByRequester retrieves all data access requests made by a specific requester.
func (r *dataAccessRepository) GetRequestsByStatus(requestAccountID, status string) ([]domain.DataAccessRequest, error) {
	var requests []domain.DataAccessRequest

	// Query the database for requests filtered by status
	err := r.db.Preload("RequesterAccount").
		Where("request_account_id = ? AND status = ?", requestAccountID, status).
		Find(&requests).Error
	if err != nil {
		return nil, err
	}

	return requests, nil
}

// UpdateRequestStatus updates the status of a data access request.
// If the status is REJECTED, the reason for rejection can also be set.
func (r *dataAccessRepository) UpdateRequestStatus(
	requestAccountID, requesterAccountID string, status constants.DataAccessRequestStatus, reasonForRejection *string,
) error {
	// Prepare the update fields
	updateData := map[string]interface{}{
		"status": status,
	}
	if status == constants.DataAccessRequestRejected && reasonForRejection != nil {
		updateData["reason_for_rejection"] = *reasonForRejection
	}

	// Update the database record with additional validation
	if err := r.db.Model(&domain.DataAccessRequest{}).
		Where("request_account_id = ? AND requester_account_id = ?", requestAccountID, requesterAccountID).
		Updates(updateData).Error; err != nil {
		return err
	}

	return nil
}
