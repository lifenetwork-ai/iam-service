package repositories

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
)

type accountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) interfaces.AccountRepository {
	return &accountRepository{db: db}
}

// Accounts
func (r *accountRepository) AccountExists(accountID string) (bool, error) {
	var account domain.Account
	err := r.db.Select("id").Where("id = ?", accountID).First(&account).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return err == nil, err
}

func (r *accountRepository) FindAccountByUsername(username string) (*domain.Account, error) {
	var account domain.Account
	if err := r.db.Where("username = ?", username).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

func (r *accountRepository) FindAccountByEmail(email string) (*domain.Account, error) {
	var account domain.Account
	if err := r.db.Where("email = ?", email).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

func (r *accountRepository) FindAccountByPhoneNumber(phone string) (*domain.Account, error) {
	var account domain.Account
	if err := r.db.Where("phone_number = ?", phone).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

func (r *accountRepository) FindAccountByAPIKey(apiKey string) (*domain.Account, error) {
	var account domain.Account
	err := r.db.Where("api_key = ?", apiKey).First(&account).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &account, err
}

func (r *accountRepository) FindAccountByID(id string) (*domain.Account, error) {
	var account domain.Account
	err := r.db.Where("id = ?", id).First(&account).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Return nil without an error if the account is not found
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("database error occurred while fetching account by ID: %w", err)
	}
	return &account, nil
}

func (r *accountRepository) CreateAccount(account *domain.Account) error {
	return r.db.Create(account).Error
}

// Data owner
func (r *accountRepository) FindDataOwnerByAccountID(accountID string) (*domain.DataOwner, error) {
	var dataOwner domain.DataOwner

	// Attempt to preload the data owner and associated account
	err := r.db.Preload("Account").Where("account_id = ?", accountID).First(&dataOwner).Error
	if err == nil {
		return &dataOwner, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Return the error if it's not a "record not found" error
		return nil, fmt.Errorf("failed to fetch data owner: %w", err)
	}

	// Use FindAccountByID to handle account fetching
	account, accErr := r.FindAccountByID(accountID)
	if accErr != nil {
		return nil, fmt.Errorf("failed to fetch account: %w", accErr)
	}

	if account == nil {
		// Return nil if the account is also not found
		return nil, nil
	}

	// Return an empty DataOwner with the account preloaded
	return &domain.DataOwner{
		AccountID: account.ID,
		Account:   *account,
	}, nil
}

func (r *accountRepository) CreateOrUpdateDataOwner(dataOwner *domain.DataOwner) error {
	existingDataOwner, err := r.FindDataOwnerByAccountID(dataOwner.AccountID)
	if err != nil {
		return err
	}
	if existingDataOwner != nil {
		dataOwner.ID = existingDataOwner.ID // Preserve the existing record's ID
	}
	return r.db.Save(dataOwner).Error
}

// Data utilizer
func (r *accountRepository) FindDataUtilizerByAccountID(accountID string) (*domain.DataUtilizer, error) {
	var dataUtilizer domain.DataUtilizer

	// Attempt to preload customer detail and associated account
	err := r.db.Preload("Account").Where("account_id = ?", accountID).First(&dataUtilizer).Error
	if err == nil {
		return &dataUtilizer, nil
	}

	// If the error is not "record not found," return it immediately
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to fetch customer detail: %w", err)
	}

	// Use FindAccountByID to check if the account exists
	account, accErr := r.FindAccountByID(accountID)
	if accErr != nil {
		return nil, fmt.Errorf("failed to fetch account: %w", accErr)
	}

	if account == nil {
		// Return nil if both the customer detail and account are not found
		return nil, nil
	}

	// Return an empty CustomerDetail with the associated account preloaded
	return &domain.DataUtilizer{
		AccountID: account.ID,
		Account:   *account,
	}, nil
}

func (r *accountRepository) CreateOrUpdateDataUtilizer(dataUtilizer *domain.DataUtilizer) error {
	existingDataUtilizer, err := r.FindDataUtilizerByAccountID(dataUtilizer.AccountID)
	if err != nil {
		return err
	}

	if existingDataUtilizer != nil {
		dataUtilizer.ID = existingDataUtilizer.ID // Preserve the existing record's ID
	}
	return r.db.Save(dataUtilizer).Error
}

// Validator detail
func (r *accountRepository) FindValidatorByAccountID(accountID string) (*domain.Validator, error) {
	var validator domain.Validator

	// Attempt to find validator-specific detail
	if err := r.db.Preload("Account").Where("account_id = ?", accountID).First(&validator).Error; err != nil {
		// If error is not "record not found," return the error immediately
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to fetch validator: %w", err)
		}

		// Handle the case where validator is not found
		account, accErr := r.FindAccountByID(accountID)
		if accErr != nil {
			return nil, fmt.Errorf("failed to fetch associated account: %w", accErr)
		}
		if account == nil {
			// Return nil if both validator detail and account are not found
			return nil, nil
		}

		// Return an empty Validator with the associated account preloaded
		return &domain.Validator{
			AccountID: account.ID,
			Account:   *account,
		}, nil
	}

	// Return the found details
	return &validator, nil
}

func (r *accountRepository) CreateOrUpdateValidator(validator *domain.Validator) error {
	existingValidator, err := r.FindValidatorByAccountID(validator.AccountID)
	if err != nil {
		return err
	}

	if existingValidator != nil {
		validator.ID = existingValidator.ID // Preserve the existing record's ID
	}
	return r.db.Save(validator).Error
}

func (r *accountRepository) FindActiveValidators() ([]domain.Validator, error) {
	var validators []domain.Validator
	err := r.db.Preload("Account").Where("is_active = ?", true).Find(&validators).Error
	if err != nil {
		return nil, err
	}
	return validators, nil
}

func (r *accountRepository) FindActiveValidatorsByIDs(validatorIDs []string) ([]domain.Validator, error) {
	var validators []domain.Validator
	err := r.db.Preload("Account").
		Where("is_active = ? AND id IN ?", true, validatorIDs).
		Find(&validators).Error
	if err != nil {
		return nil, err
	}
	return validators, nil
}

// UpdateAccount updates the details of an existing account.
func (r *accountRepository) UpdateAccount(account *domain.Account) error {
	// Use Save to update the account, it will update all fields of the account struct.
	if err := r.db.Save(account).Error; err != nil {
		return err
	}
	return nil
}
