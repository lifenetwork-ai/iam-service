package repositories

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
)

type userIdentityRepository struct {
	db *gorm.DB
}

func NewUserIdentityRepository(db *gorm.DB) domainrepo.UserIdentityRepository {
	return &userIdentityRepository{db: db}
}

func (r *userIdentityRepository) GetByID(ctx context.Context, tx *gorm.DB, identityID string) (*domain.UserIdentity, error) {
	var identity domain.UserIdentity
	db := r.db
	if tx != nil {
		db = tx
	}
	err := db.WithContext(ctx).First(&identity).Error
	if err != nil {
		return nil, err
	}
	return &identity, nil
}

func (r *userIdentityRepository) GetByTypeAndValue(
	ctx context.Context,
	tx *gorm.DB,
	tenantID string,
	identityType string,
	value string,
) (*domain.UserIdentity, error) {
	var identity domain.UserIdentity
	db := r.db
	if tx != nil {
		db = tx
	}
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND type = ? AND value = ?",
			tenantID, identityType, value).
		First(&identity).Error
	if err != nil {
		return nil, err
	}
	return &identity, nil
}

func (r *userIdentityRepository) FindGlobalUserIDByIdentity(
	ctx context.Context,
	tenantID string,
	identityType string,
	value string,
) (string, error) {
	var out struct {
		GlobalUserID string
	}
	err := r.db.WithContext(ctx).
		Model(&domain.UserIdentity{}).
		Select("global_user_id").
		Where("tenant_id = ? AND type = ? AND value = ?",
			tenantID, identityType, value).
		First(&out).Error
	if err != nil {
		return "", err
	}
	return out.GlobalUserID, nil
}

func (r *userIdentityRepository) InsertOnceByTenantUserAndType(
	ctx context.Context,
	tx *gorm.DB,
	tenantID string,
	globalUserID string,
	idType, value string,
) (bool, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	rec := &domain.UserIdentity{
		TenantID:     tenantID,
		GlobalUserID: globalUserID,
		Type:         idType,
		Value:        value,
	}
	res := db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "tenant_id"}, {Name: "global_user_id"}, {Name: "type"}},
			DoNothing: true,
		}).
		Create(rec)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}

func (r *userIdentityRepository) Update(tx *gorm.DB, identity *domain.UserIdentity) error {
	return tx.Updates(identity).Error
}

// ExistsWithinTenant checks if an identity exists within a tenant
func (r *userIdentityRepository) ExistsWithinTenant(
	ctx context.Context,
	tenantID, identityType, value string,
) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.UserIdentity{}).
		Where("tenant_id = ? AND type = ? AND value = ?",
			tenantID, identityType, value).
		Count(&count).Error
	return count > 0, err
}

// GetByTenantAndTenantUserID retrieves the first user identity by tenant ID and tenant user ID
// One tenant user can have multiple user identities, but we only return the first one
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

// ExistsByTenantGlobalUserIDAndType checks if a user identity exists by tenant ID, global user ID, and type
func (r *userIdentityRepository) ExistsByTenantGlobalUserIDAndType(
	ctx context.Context,
	tenantID, globalUserID, identityType string,
) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.UserIdentity{}).
		Where("tenant_id = ? AND global_user_id = ? AND type = ?",
			tenantID, globalUserID, identityType).
		Count(&count).Error
	return count > 0, err
}

func (r *userIdentityRepository) GetByGlobalUserID(
	ctx context.Context,
	tx *gorm.DB,
	tenantID, globalUserID string,
) ([]domain.UserIdentity, error) {
	var identities []domain.UserIdentity

	db := r.db
	if tx != nil {
		db = tx
	}

	err := db.WithContext(ctx).
		Where("tenant_id = ? AND global_user_id = ?", tenantID, globalUserID).
		Find(&identities).Error
	if err != nil {
		return nil, err
	}
	return identities, nil
}

func (r *userIdentityRepository) Delete(tx *gorm.DB, identityID string) error {
	return tx.Delete(&domain.UserIdentity{ID: identityID}).Error
}
