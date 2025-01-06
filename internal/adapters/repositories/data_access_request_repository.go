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

// GetPendingRequests retrieves a list of pending data access requests for a specific user.
func (r *dataAccessRepository) GetPendingRequests(userID string) ([]domain.DataAccessRequest, error) {
	var requests []domain.DataAccessRequest

	// Query the database for pending requests where the user is the recipient
	err := r.db.Where("user_id = ? AND status = ?", userID, string(constants.DataAccessRequestPending)).
		Find(&requests).Error
	if err != nil {
		return nil, err
	}

	return requests, nil
}
