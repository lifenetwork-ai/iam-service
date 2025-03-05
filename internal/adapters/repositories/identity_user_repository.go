package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	cachingtypes "github.com/genefriendway/human-network-iam/infrastructures/caching/types"
	infrainterfaces "github.com/genefriendway/human-network-iam/infrastructures/interfaces"
	interfaces "github.com/genefriendway/human-network-iam/internal/adapters/repositories/types"
	entities "github.com/genefriendway/human-network-iam/internal/domain/entities"
	"github.com/genefriendway/human-network-iam/packages/logger"
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

func (r *identityRepository) FindByID(
	ctx context.Context,
	userID string,
) (*entities.IdentityUser, error) {
	if userID == "" {
		return nil, nil
	}

	organizationId := ctx.Value("organizationId").(string)
	if organizationId == "" {
		return nil, fmt.Errorf("missing organization ID")
	}

	cacheKey := &cachingtypes.Keyer{
		Raw: fmt.Sprintf("identity_user_%s", userID),
	}

	// Find in cache, if not found, find in database
	var cacheRequester interface{}
	err := r.cache.RetrieveItem(cacheKey, &cacheRequester)
	if err == nil {
		if user, ok := cacheRequester.(entities.IdentityUser); ok {
			return &user, nil
		}
	}

	var entity entities.IdentityUser
	err = r.db.Where("organization_id = ? AND id = ?", organizationId, userID).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err == nil {
		err := r.cache.SaveItem(cacheKey, entity, 1*time.Hour)
		if err != nil {
			logger.GetLogger().Errorf("Failed to save cache: %v", err)
		}
	}

	return &entity, err
}

func (r *identityRepository) FindByPhone(
	ctx context.Context,
	phone string,
) (*entities.IdentityUser, error) {
	if phone == "" {
		return nil, nil
	}

	organizationId := ctx.Value("organizationId").(string)
	if organizationId == "" {
		return nil, fmt.Errorf("missing organization ID")
	}

	var entity entities.IdentityUser
	err := r.db.Where("organization_id = ? AND phone = ?", organizationId, phone).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &entity, err
}

func (r *identityRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*entities.IdentityUser, error) {
	if email == "" {
		return nil, nil
	}

	organizationId := ctx.Value("organizationId").(string)
	if organizationId == "" {
		return nil, fmt.Errorf("missing organization ID")
	}

	var entity entities.IdentityUser
	err := r.db.Where("organization_id = ? AND email = ?", organizationId, email).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &entity, err
}

func (r *identityRepository) FindByUsername(
	ctx context.Context,
	username string,
) (*entities.IdentityUser, error) {
	if username == "" {
		return nil, nil
	}

	organizationId := ctx.Value("organizationId").(string)
	if organizationId == "" {
		return nil, fmt.Errorf("missing organization ID")
	}

	var entity entities.IdentityUser
	err := r.db.Where("organization_id = ? AND username = ?", organizationId, username).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &entity, err
}

func (r *identityRepository) FindBySelfAuthenticateID(
	ctx context.Context,
	selfAuthID string,
) (*entities.IdentityUser, error) {
	if selfAuthID == "" {
		return nil, nil
	}

	organizationId := ctx.Value("organizationId").(string)
	if organizationId == "" {
		return nil, fmt.Errorf("missing organization ID")
	}

	var entity entities.IdentityUser
	err := r.db.Where("organization_id = ? AND self_authenticate_id = ?", organizationId, selfAuthID).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &entity, err
}

func (r *identityRepository) FindByGoogleID(
	ctx context.Context,
	googleID string,
) (*entities.IdentityUser, error) {
	if googleID == "" {
		return nil, nil
	}

	organizationId := ctx.Value("organizationId").(string)
	if organizationId == "" {
		return nil, fmt.Errorf("missing organization ID")
	}

	var entity entities.IdentityUser
	err := r.db.Where("organization_id = ? AND google_id = ?", organizationId, googleID).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &entity, err
}

func (r *identityRepository) FindByFacebookID(
	ctx context.Context,
	facebookID string,
) (*entities.IdentityUser, error) {
	if facebookID == "" {
		return nil, nil
	}

	organizationId := ctx.Value("organizationId").(string)
	if organizationId == "" {
		return nil, fmt.Errorf("missing organization ID")
	}

	var entity entities.IdentityUser
	err := r.db.Where("organization_id = ? AND facebook_id = ?", organizationId, facebookID).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &entity, err
}

func (r *identityRepository) FindByAppleID(
	ctx context.Context,
	appleID string,
) (*entities.IdentityUser, error) {
	if appleID == "" {
		return nil, nil
	}

	organizationId := ctx.Value("organizationId").(string)
	if organizationId == "" {
		return nil, fmt.Errorf("missing organization ID")
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

	if err == nil {
		// Save to cache for 1 hour
		cacheKey := &cachingtypes.Keyer{
			Raw: fmt.Sprintf("identity_user_%s", entity.ID),
		}
		cacheErr := r.cache.SaveItem(cacheKey, entity, 1*time.Hour)
		if cacheErr != nil {
			logger.GetLogger().Errorf("Failed to remove cache: %v", cacheErr)
		}
	}

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

	if err == nil {
		// Save to cache for 1 hour
		cacheKey := &cachingtypes.Keyer{
			Raw: fmt.Sprintf("identity_user_%s", entity.ID),
		}
		cacheErr := r.cache.SaveItem(cacheKey, entity, 1*time.Hour)
		if cacheErr != nil {
			logger.GetLogger().Errorf("Failed to remove cache: %v", cacheErr)
		}
	}

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

	if err == nil {
		// Remove from cache
		cacheKey := &cachingtypes.Keyer{
			Raw: fmt.Sprintf("identity_user_%s", userID),
		}
		cacheErr := r.cache.RemoveItem(cacheKey)
		if cacheErr != nil {
			logger.GetLogger().Errorf("Failed to remove cache: %v", cacheErr)
		}
	}

	return err
}
