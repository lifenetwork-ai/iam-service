package repositories

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	infrainterfaces "github.com/genefriendway/human-network-iam/infrastructures/interfaces"
	interfaces "github.com/genefriendway/human-network-iam/internal/adapters/repositories/types"
	entities "github.com/genefriendway/human-network-iam/internal/domain/entities"
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
) (*entities.IdentityUser, error) {
	organizationId := ctx.Value("organizationId").(string)
	if organizationId == "" {
		return nil, fmt.Errorf("missing organization ID")
	}

	if phone == "" {
		return nil, nil
	}

	var entity entities.IdentityUser
	err := r.db.Where("organization_id = ? AND phone = ?", organizationId, phone).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &entity, err
}

func (r *identityRepository) GetByEmail(
	ctx context.Context,
	email string,
) (*entities.IdentityUser, error) {
	organizationId := ctx.Value("organizationId").(string)
	if organizationId == "" {
		return nil, fmt.Errorf("missing organization ID")
	}

	if email == "" {
		return nil, nil
	}

	var entity entities.IdentityUser
	err := r.db.Where("organization_id = ? AND email = ?", organizationId, email).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &entity, err
}

func (r *identityRepository) GetByUsername(
	ctx context.Context,
	username string,
) (*entities.IdentityUser, error) {
	organizationId := ctx.Value("organizationId").(string)
	if organizationId == "" {
		return nil, fmt.Errorf("missing organization ID")
	}

	if username == "" {
		return nil, nil
	}

	var entity entities.IdentityUser
	err := r.db.Where("organization_id = ? AND username = ?", organizationId, username).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &entity, err
}

func (r *identityRepository) GetByLifeAIID(
	ctx context.Context,
	lifeAIID string,
) (*entities.IdentityUser, error) {
	organizationId := ctx.Value("organizationId").(string)
	if organizationId == "" {
		return nil, fmt.Errorf("missing organization ID")
	}

	if lifeAIID == "" {
		return nil, nil
	}

	var entity entities.IdentityUser
	err := r.db.Where("organization_id = ? AND lifeai_id = ?", organizationId, lifeAIID).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &entity, err
}

func (r *identityRepository) GetByGoogleID(
	ctx context.Context,
	googleID string,
) (*entities.IdentityUser, error) {
	organizationId := ctx.Value("organizationId").(string)
	if organizationId == "" {
		return nil, fmt.Errorf("missing organization ID")
	}

	if googleID == "" {
		return nil, nil
	}

	var entity entities.IdentityUser
	err := r.db.Where("organization_id = ? AND google_id = ?", organizationId, googleID).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &entity, err
}

func (r *identityRepository) GetByFacebookID(
	ctx context.Context,
	facebookID string,
) (*entities.IdentityUser, error) {
	organizationId := ctx.Value("organizationId").(string)
	if organizationId == "" {
		return nil, fmt.Errorf("missing organization ID")
	}

	if facebookID == "" {
		return nil, nil
	}

	var entity entities.IdentityUser
	err := r.db.Where("organization_id = ? AND facebook_id = ?", organizationId, facebookID).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &entity, err
}

func (r *identityRepository) GetByAppleID(
	ctx context.Context,
	appleID string,
) (*entities.IdentityUser, error) {
	organizationId := ctx.Value("organizationId").(string)
	if organizationId == "" {
		return nil, fmt.Errorf("missing organization ID")
	}

	if appleID == "" {
		return nil, nil
	}

	var entity entities.IdentityUser
	err := r.db.Where("organization_id = ? AND apple_id = ?", organizationId, appleID).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &entity, err
}

func (r *identityRepository) Create(
	ctx context.Context,
	entity *entities.IdentityUser,
) error {
	organizationId := ctx.Value("organizationId").(string)
	if organizationId == "" {
		return fmt.Errorf("missing organization ID")
	}

	entity.OrganizationId = organizationId
	err := r.db.Create(entity).Error
	return err
}

func (r *identityRepository) Update(
	ctx context.Context,
	entity *entities.IdentityUser,
) error {
	organizationId := ctx.Value("organizationId").(string)
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
	organizationId := ctx.Value("organizationId").(string)
	if organizationId == "" {
		return fmt.Errorf("missing organization ID")
	}

	err := r.db.Where("organization_id = ? AND id = ?", organizationId, userID).Delete(&entities.IdentityUser{}).Error
	return err
}
