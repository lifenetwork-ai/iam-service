package identity_user

import (
	"gorm.io/gorm"

	"github.com/genefriendway/human-network-iam/internal/dto"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type identityRepository struct {
	db *gorm.DB
}

func NewIdentityUserRepository(db *gorm.DB) interfaces.IdentityUserRepository {
	return &identityRepository{db: db}
}

func (r *identityRepository) GetByPhone(phone string) (*dto.IdentityUserDTO, error) {
	var entity dto.IdentityUserDTO
	err := r.db.Where("phone = ?", phone).First(&entity).Error
	return &entity, err
}

func (r *identityRepository) GetByEmail(email string) (*dto.IdentityUserDTO, error) {
	var entity dto.IdentityUserDTO
	err := r.db.Where("email = ?", email).First(&entity).Error
	return &entity, err
}

func (r *identityRepository) GetByUsername(username string) (*dto.IdentityUserDTO, error) {
	var entity dto.IdentityUserDTO
	err := r.db.Where("username = ?", username).First(&entity).Error
	return &entity, err
}

func (r *identityRepository) GetByGoogleID(googleID string) (*dto.IdentityUserDTO, error) {
	var entity dto.IdentityUserDTO
	err := r.db.Where("google_id = ?", googleID).First(&entity).Error
	return &entity, err
}

func (r *identityRepository) GetByFacebookID(facebookID string) (*dto.IdentityUserDTO, error) {
	var entity dto.IdentityUserDTO
	err := r.db.Where("facebook_id = ?", facebookID).First(&entity).Error
	return &entity, err
}

func (r *identityRepository) GetByAppleID(appleID string) (*dto.IdentityUserDTO, error) {
	var entity dto.IdentityUserDTO
	err := r.db.Where("apple_id = ?", appleID).First(&entity).Error
	return &entity, err
}

func (r *identityRepository) Create(entity *dto.IdentityUserDTO) error {
	err := r.db.Create(entity).Error
	return err
}

func (r *identityRepository) Update(entity *dto.IdentityUserDTO) error {
	err := r.db.Save(entity).Error
	return err
}

func (r *identityRepository) Delete(userID string) error {
	err := r.db.Where("id = ?", userID).Delete(&dto.IdentityUserDTO{}).Error
	return err
}
