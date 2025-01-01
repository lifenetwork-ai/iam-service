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

func (r *accountRepository) FindAccountByID(id uint64) (*domain.Account, error) {
	var account domain.Account
	if err := r.db.First(&account, id).Error; err != nil {
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
func (r *accountRepository) FindUserDetailByAccountID(accountID uint64) (*domain.UserDetail, error) {
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

func (r *accountRepository) CreateUserDetail(detail *domain.UserDetail) error {
	return r.db.Create(detail).Error
}

// Partner detail
func (r *accountRepository) FindPartnerDetailByAccountID(accountID uint64) (*domain.PartnerDetail, error) {
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

func (r *accountRepository) CreatePartnerDetail(detail *domain.PartnerDetail) error {
	return r.db.Create(detail).Error
}

// Customer detail
func (r *accountRepository) FindCustomerDetailByAccountID(accountID uint64) (*domain.CustomerDetail, error) {
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

func (r *accountRepository) CreateCustomerDetail(detail *domain.CustomerDetail) error {
	return r.db.Create(detail).Error
}

// Validator detail
func (r *accountRepository) FindValidatorDetailByAccountID(accountID uint64) (*domain.ValidatorDetail, error) {
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

func (r *accountRepository) CreateValidatorDetail(detail *domain.ValidatorDetail) error {
	return r.db.Create(detail).Error
}
