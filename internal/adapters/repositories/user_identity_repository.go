package repositories

import (
	"context"

	"gorm.io/gorm"

	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
)

type userIdentityRepository struct {
	db *gorm.DB
}

func NewUserIdentityRepository(db *gorm.DB) domainrepo.UserIdentityRepository {
	return &userIdentityRepository{db: db}
}

func (r *userIdentityRepository) GetByGlobalUserID(ctx context.Context, globalUserID string) ([]domain.UserIdentity, error) {
	var list []domain.UserIdentity
	if err := r.db.WithContext(ctx).Where("global_user_id = ?", globalUserID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *userIdentityRepository) GetByTypeAndValue(
	ctx context.Context,
	tx *gorm.DB,
	identityType string,
	value string,
) (*domain.UserIdentity, error) {
	var identity domain.UserIdentity

	db := r.db.WithContext(ctx)
	if tx != nil {
		db = tx.WithContext(ctx)
	}

	if err := db.
		Where("type = ? AND value = ?", identityType, value).
		First(&identity).Error; err != nil {
		return nil, err
	}

	return &identity, nil
}

func (r *userIdentityRepository) FindGlobalUserIDByIdentity(ctx context.Context, identityType, value string) (string, error) {
	var identity domain.UserIdentity
	if err := r.db.WithContext(ctx).
		Select("global_user_id").
		Where("type = ? AND value = ?", identityType, value).
		First(&identity).Error; err != nil {
		return "", err
	}
	return identity.GlobalUserID, nil
}

func (r *userIdentityRepository) FirstOrCreate(tx *gorm.DB, identity *domain.UserIdentity) error {
	return tx.FirstOrCreate(identity, domain.UserIdentity{
		Type:  identity.Type,
		Value: identity.Value,
	}).Error
}

func (r *userIdentityRepository) Update(tx *gorm.DB, identity *domain.UserIdentity) error {
	return tx.Save(identity).Error
}

// ExistsWithinTenant checks if an identity exists within a tenant
func (r *userIdentityRepository) ExistsWithinTenant(ctx context.Context, tenantID, identityType, value string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.UserIdentity{}).
		Joins("JOIN global_users ON global_users.id = user_identities.global_user_id").
		Joins("JOIN user_identifier_mapping ON user_identifier_mapping.global_user_id = global_users.id").
		Where("user_identifier_mapping.tenant_id = ? AND user_identities.type = ? AND user_identities.value = ?", tenantID, identityType, value).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetByTenantAndTenantUserID retrieves a single user identity by tenant ID and tenant user ID
func (r *userIdentityRepository) GetByTenantAndTenantUserID(
	ctx context.Context,
	tx *gorm.DB,
	tenantID, tenantUserID string,
) (*domain.UserIdentity, error) {
	var identity domain.UserIdentity

	db := r.db.WithContext(ctx)
	if tx != nil {
		db = tx.WithContext(ctx)
	}

	err := db.
		Model(&domain.UserIdentity{}).
		Joins("JOIN user_identifier_mapping ON user_identifier_mapping.global_user_id = user_identities.global_user_id").
		Where("user_identifier_mapping.tenant_id = ? AND user_identifier_mapping.tenant_user_id = ?", tenantID, tenantUserID).
		First(&identity).Error
	if err != nil {
		return nil, err
	}
	return &identity, nil
}

// ExistsByGlobalUserIDAndType checks if a user identity exists by global user ID and type
func (r *userIdentityRepository) ExistsByGlobalUserIDAndType(ctx context.Context, globalUserID, identityType string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.UserIdentity{}).
		Where("global_user_id = ? AND type = ?", globalUserID, identityType).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
