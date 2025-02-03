package repositories

import (
	"gorm.io/gorm"

	"github.com/genefriendway/human-network-iam/internal/dto"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type identityRepository struct {
	db *gorm.DB
}

func NewIdentityRepository(db *gorm.DB) interfaces.IdentityRepository {
	return &identityRepository{db: db}
}

func (r *identityRepository) GetAccountByPhone(phone string) (*dto.AccountDTO, error) {
	var account dto.AccountDTO
	err := r.db.Where("phone = ?", phone).First(&account).Error
	return &account, err
}

func (r *identityRepository) GetAccountByEmail(email string) (*dto.AccountDTO, error) {
	var account dto.AccountDTO
	err := r.db.Where("email = ?", email).First(&account).Error
	return &account, err
}

func (r *identityRepository) GetAccountByUsername(username string) (*dto.AccountDTO, error) {
	var account dto.AccountDTO
	err := r.db.Where("username = ?", username).First(&account).Error
	return &account, err
}

func (r *identityRepository) GetAccountByGoogleID(googleID string) (*dto.AccountDTO, error) {
	var account dto.AccountDTO
	err := r.db.Where("google_id = ?", googleID).First(&account).Error
	return &account, err
}

func (r *identityRepository) GetAccountByFacebookID(facebookID string) (*dto.AccountDTO, error) {
	var account dto.AccountDTO
	err := r.db.Where("facebook_id = ?", facebookID).First(&account).Error
	return &account, err
}

func (r *identityRepository) GetAccountByAppleID(appleID string) (*dto.AccountDTO, error) {
	var account dto.AccountDTO
	err := r.db.Where("apple_id = ?", appleID).First(&account).Error
	return &account, err
}

func (r *identityRepository) CreateAccount(account *dto.AccountDTO) error {
	err := r.db.Create(account).Error
	return err
}

func (r *identityRepository) UpdateAccount(account *dto.AccountDTO) error {
	err := r.db.Save(account).Error
	return err
}

func (r *identityRepository) DeleteAccount(userID string) error {
	err := r.db.Where("id = ?", userID).Delete(&dto.AccountDTO{}).Error
	return err
}
