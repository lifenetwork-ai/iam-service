package repositories

import (
	"errors"

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

func (r *accountRepository) FindAccountByID(id string) (*domain.Account, error) {
	var account domain.Account
	if err := r.db.Where("id = ?", id).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

func (r *accountRepository) CreateAccount(account *domain.Account) error {
	return r.db.Create(account).Error
}

// User detail
func (r *accountRepository) FindUserDetailByAccountID(accountID string) (*domain.UserDetail, error) {
	var details domain.UserDetail
	// Use Preload to eagerly load the Account relationship
	if err := r.db.Preload("Account").Where("account_id = ?", accountID).First(&details).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &details, nil
}

func (r *accountRepository) CreateOrUpdateUserDetail(detail *domain.UserDetail) error {
	existingDetail, err := r.FindUserDetailByAccountID(detail.AccountID)
	if err != nil {
		return err
	}

	if existingDetail != nil {
		detail.ID = existingDetail.ID // Preserve the existing record's ID
	}
	return r.db.Save(detail).Error
}

// Partner detail
func (r *accountRepository) FindPartnerDetailByAccountID(accountID string) (*domain.PartnerDetail, error) {
	var details domain.PartnerDetail
	// Use Preload to eagerly load the Account relationship
	if err := r.db.Preload("Account").Where("account_id = ?", accountID).First(&details).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &details, nil
}

func (r *accountRepository) CreateOrUpdatePartnerDetail(detail *domain.PartnerDetail) error {
	existingDetail, err := r.FindPartnerDetailByAccountID(detail.AccountID)
	if err != nil {
		return err
	}

	if existingDetail != nil {
		detail.ID = existingDetail.ID // Preserve the existing record's ID
	}
	return r.db.Save(detail).Error
}

// Customer detail
func (r *accountRepository) FindCustomerDetailByAccountID(accountID string) (*domain.CustomerDetail, error) {
	var details domain.CustomerDetail
	// Use Preload to eagerly load the Account relationship
	if err := r.db.Preload("Account").Where("account_id = ?", accountID).First(&details).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &details, nil
}

func (r *accountRepository) CreateOrUpdateCustomerDetail(detail *domain.CustomerDetail) error {
	existingDetail, err := r.FindCustomerDetailByAccountID(detail.AccountID)
	if err != nil {
		return err
	}

	if existingDetail != nil {
		detail.ID = existingDetail.ID // Preserve the existing record's ID
	}
	return r.db.Save(detail).Error
}

// Validator detail
func (r *accountRepository) FindValidatorDetailByAccountID(accountID string) (*domain.ValidatorDetail, error) {
	var details domain.ValidatorDetail
	// Use Preload to eagerly load the Account relationship
	if err := r.db.Preload("Account").Where("account_id = ?", accountID).First(&details).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &details, nil
}

func (r *accountRepository) CreateOrUpdateValidatorDetail(detail *domain.ValidatorDetail) error {
	existingDetail, err := r.FindValidatorDetailByAccountID(detail.AccountID)
	if err != nil {
		return err
	}

	if existingDetail != nil {
		detail.ID = existingDetail.ID // Preserve the existing record's ID
	}
	return r.db.Save(detail).Error
}

func (r *accountRepository) FindActiveValidators() ([]domain.ValidatorDetail, error) {
	var validators []domain.ValidatorDetail
	err := r.db.Where("is_active = ?", true).Find(&validators).Error
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
