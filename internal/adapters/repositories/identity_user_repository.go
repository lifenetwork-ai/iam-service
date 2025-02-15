package repositories

import (
	"context"

	"gorm.io/gorm"

	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/dto"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type identityRepository struct {
	db *gorm.DB
}

func NewIdentityUserRepository(db *gorm.DB) interfaces.IdentityUserRepository {
	return &identityRepository{db: db}
}

func (r *identityRepository) GetByPhone(
	ctx context.Context,
	organizationId string,
	phone string,
) (*dto.IdentityUserDTO, error) {
	var entity dto.IdentityUserDTO
	err := r.db.Where("phone = ?", phone).First(&entity).Error
	return &entity, err
}

func (r *identityRepository) GetByEmail(
	ctx context.Context,
	organizationId string,
	email string,
) (*dto.IdentityUserDTO, error) {
	var entity dto.IdentityUserDTO
	err := r.db.Where("email = ?", email).First(&entity).Error
	return &entity, err
}

func (r *identityRepository) GetByUsername(
	ctx context.Context,
	organizationId string,
	username string,
) (*dto.IdentityUserDTO, error) {
	var entity dto.IdentityUserDTO
	err := r.db.Where("username = ?", username).First(&entity).Error
	return &entity, err
}

func (r *identityRepository) GetByGoogleID(
	ctx context.Context,
	organizationId string,
	googleID string,
) (*dto.IdentityUserDTO, error) {
	var entity dto.IdentityUserDTO
	err := r.db.Where("google_id = ?", googleID).First(&entity).Error
	return &entity, err
}

func (r *identityRepository) GetByFacebookID(
	ctx context.Context,
	organizationId string,
	facebookID string,
) (*dto.IdentityUserDTO, error) {
	var entity dto.IdentityUserDTO
	err := r.db.Where("facebook_id = ?", facebookID).First(&entity).Error
	return &entity, err
}

func (r *identityRepository) GetByAppleID(
	ctx context.Context,
	organizationId string,
	appleID string,
) (*dto.IdentityUserDTO, error) {
	var entity dto.IdentityUserDTO
	err := r.db.Where("apple_id = ?", appleID).First(&entity).Error
	return &entity, err
}

func (r *identityRepository) Create(
	ctx context.Context,
	entity *domain.IdentityUser,
) error {
	err := r.db.Create(entity).Error
	return err
}

func (r *identityRepository) Update(
	ctx context.Context,
	entity *domain.IdentityUser,
) error {
	err := r.db.Save(entity).Error
	return err
}

func (r *identityRepository) Delete(
	ctx context.Context,
	userID string,
) error {
	err := r.db.Where("id = ?", userID).Delete(&domain.IdentityUser{}).Error
	return err
}
