package repositories

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	infrainterfaces "github.com/genefriendway/human-network-iam/infrastructures/interfaces"
	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type identityRepository struct {
	db    *gorm.DB
	cache infrainterfaces.CacheRepository
}

func NewIdentityUserRepository(
	db *gorm.DB,
	cache infrainterfaces.CacheRepository,
) interfaces.IdentityUserRepository {
	return &identityRepository{
		db:    db,
		cache: cache,
	}
}

func (r *identityRepository) GetByPhone(
	ctx context.Context,
	phone string,
) (*domain.IdentityUser, error) {
	organizationId := ctx.Value("organization_id").(string)
	if organizationId == "" {
		return nil, fmt.Errorf("missing organization ID")
	}

	var entity domain.IdentityUser
	err := r.db.Where("organization_id = ? AND phone = ?", organizationId, phone).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &entity, err
}

func (r *identityRepository) GetByEmail(
	ctx context.Context,
	email string,
) (*domain.IdentityUser, error) {
	organizationId := ctx.Value("organization_id").(string)
	if organizationId == "" {
		return nil, fmt.Errorf("missing organization ID")
	}

	var entity domain.IdentityUser
	err := r.db.Where("organization_id = ? AND email = ?", organizationId, email).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &entity, err
}

func (r *identityRepository) GetByUsername(
	ctx context.Context,
	username string,
) (*domain.IdentityUser, error) {
	organizationId := ctx.Value("organization_id").(string)
	if organizationId == "" {
		return nil, fmt.Errorf("missing organization ID")
	}

	var entity domain.IdentityUser
	err := r.db.Where("organization_id = ? AND username = ?", organizationId, username).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &entity, err
}

func (r *identityRepository) GetByGoogleID(
	ctx context.Context,
	googleID string,
) (*domain.IdentityUser, error) {
	organizationId := ctx.Value("organization_id").(string)
	if organizationId == "" {
		return nil, fmt.Errorf("missing organization ID")
	}

	var entity domain.IdentityUser
	err := r.db.Where("organization_id = ? AND google_id = ?", organizationId, googleID).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &entity, err
}

func (r *identityRepository) GetByFacebookID(
	ctx context.Context,
	facebookID string,
) (*domain.IdentityUser, error) {
	organizationId := ctx.Value("organization_id").(string)
	if organizationId == "" {
		return nil, fmt.Errorf("missing organization ID")
	}

	var entity domain.IdentityUser
	err := r.db.Where("organization_id = ? AND facebook_id = ?", organizationId, facebookID).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &entity, err
}

func (r *identityRepository) GetByAppleID(
	ctx context.Context,
	appleID string,
) (*domain.IdentityUser, error) {
	organizationId := ctx.Value("organization_id").(string)
	if organizationId == "" {
		return nil, fmt.Errorf("missing organization ID")
	}

	var entity domain.IdentityUser
	err := r.db.Where("organization_id = ? AND apple_id = ?", organizationId, appleID).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &entity, err
}

func (r *identityRepository) Create(
	ctx context.Context,
	entity *domain.IdentityUser,
) error {
	organizationId := ctx.Value("organization_id").(string)
	if organizationId == "" {
		return fmt.Errorf("missing organization ID")
	}

	entity.OrganizationId = organizationId
	err := r.db.Create(entity).Error
	return err
}

func (r *identityRepository) Update(
	ctx context.Context,
	entity *domain.IdentityUser,
) error {
	organizationId := ctx.Value("organization_id").(string)
	if organizationId == "" {
		return fmt.Errorf("missing organization ID")
	}

	entity.OrganizationId = organizationId
	err := r.db.Save(entity).Error
	return err
}

func (r *identityRepository) Delete(
	ctx context.Context,
	userID string,
) error {
	organizationId := ctx.Value("organization_id").(string)
	if organizationId == "" {
		return fmt.Errorf("missing organization ID")
	}

	err := r.db.Where("organization_id = ? AND id = ?", organizationId, userID).Delete(&domain.IdentityUser{}).Error
	return err
}
