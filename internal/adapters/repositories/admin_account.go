package repositories

import (
	"gorm.io/gorm"

	"github.com/google/uuid"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
)

type adminAccountRepository struct {
	db *gorm.DB
}

// NewAdminAccountRepository creates a new admin account repository
func NewAdminAccountRepository(db *gorm.DB) domainrepo.AdminAccountRepository {
	return &adminAccountRepository{
		db: db,
	}
}

// Create creates a new admin account
func (r *adminAccountRepository) Create(account *domain.AdminAccount) error {
	return r.db.Create(account).Error
}

// GetByUsername retrieves an admin account by username
func (r *adminAccountRepository) GetByUsername(username string) (*domain.AdminAccount, error) {
	var account domain.AdminAccount
	err := r.db.Where("username = ?", username).First(&account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

// GetByID retrieves an admin account by ID
func (r *adminAccountRepository) GetByID(id string) (*domain.AdminAccount, error) {
	var account domain.AdminAccount
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	err = r.db.Where("id = ?", parsedID).First(&account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

// Update updates an admin account
func (r *adminAccountRepository) Update(account *domain.AdminAccount) error {
	return r.db.Save(account).Error
}

// Delete deletes an admin account
func (r *adminAccountRepository) Delete(id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return r.db.Delete(&domain.AdminAccount{}, "id = ?", parsedID).Error
}
