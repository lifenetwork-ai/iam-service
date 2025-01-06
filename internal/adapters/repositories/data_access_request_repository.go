package repositories

import (
	"gorm.io/gorm"

	"github.com/genefriendway/human-network-auth/internal/domain"
)

type dataAccessRequestRepository struct {
	db *gorm.DB
}

func NewDataAccessRequestRepository(db *gorm.DB) *dataAccessRequestRepository {
	return &dataAccessRequestRepository{db: db}
}

// CreateDataAccessRequest creates a new data access request in the database.
func (r *dataAccessRequestRepository) CreateDataAccessRequest(request *domain.DataAccessRequest) error {
	// Use GORM's Create method to insert a new data access request
	if err := r.db.Create(request).Error; err != nil {
		return err
	}
	return nil
}
