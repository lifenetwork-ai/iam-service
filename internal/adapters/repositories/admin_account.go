package repositories

import (
	"github.com/google/uuid"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	"gorm.io/gorm"
)

type adminAccountRepository struct {
	db *gorm.DB
}

// NewAdminAccountRepository creates a new admin account repository
func NewAdminAccountRepository(db *gorm.DB) interfaces.AdminAccountRepository {
	return &adminAccountRepository{
		db: db,
	}
}

// Create creates a new admin account
func (r *adminAccountRepository) Create(db *gorm.DB, account *domain.AdminAccount) error {
	return db.Create(account).Error
}

// GetByEmail retrieves an admin account by email
func (r *adminAccountRepository) GetByEmail(email string) (*domain.AdminAccount, error) {
	var account domain.AdminAccount
	err := r.db.Where("email = ?", email).First(&account).Error
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
func (r *adminAccountRepository) Update(db *gorm.DB, account *domain.AdminAccount) error {
	return db.Save(account).Error
}

// Delete deletes an admin account
func (r *adminAccountRepository) Delete(db *gorm.DB, id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return db.Delete(&domain.AdminAccount{}, "id = ?", parsedID).Error
}
