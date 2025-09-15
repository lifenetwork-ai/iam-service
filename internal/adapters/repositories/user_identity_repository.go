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
	if err := db.WithContext(ctx).
		Where("tenant_id = ? AND type = ? AND value = ?", tenantID, identityType, value).
		First(&identity).Error; err != nil {
		return nil, err
	}
	return &identity, nil
}

func (r *userIdentityRepository) InsertOnceByKratosUserAndType(
	ctx context.Context,
	tx *gorm.DB,
	tenantID string,
	kratosUserID string,
	globalUserID string,
	idType string,
	value string,
) (bool, error) {
	db := r.db
	if tx != nil {
		db = tx
	}

	rec := &domain.UserIdentity{
		TenantID:     tenantID,
		KratosUserID: kratosUserID,
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
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Updates(identity).Error
}

// ExistsWithinTenant checks if an identity exists within a tenant (by type+value).
func (r *userIdentityRepository) ExistsWithinTenant(
	ctx context.Context,
	tenantID, identityType, value string,
) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.UserIdentity{}).
		Where("tenant_id = ? AND type = ? AND value = ?", tenantID, identityType, value).
		Count(&count).Error
	return count > 0, err
}

// GetByTenantAndKratosUserID retrieves identities by (tenant_id, kratos_user_id).
func (r *userIdentityRepository) GetByTenantAndKratosUserID(
	ctx context.Context,
	tx *gorm.DB,
	tenantID, kratosUserID string,
) (*domain.UserIdentity, error) {
	db := r.db
	if tx != nil {
		db = tx
	}

	var identity *domain.UserIdentity
	err := db.WithContext(ctx).
		Where(`
			tenant_id = ? 
			AND global_user_id = (
				SELECT global_user_id 
				FROM user_identities 
				WHERE tenant_id = ? AND kratos_user_id = ? 
				LIMIT 1
			)
		`, tenantID, tenantID, kratosUserID).
		First(&identity).Error
	if err != nil {
		return nil, err
	}

	return identity, nil
}

// ListByTenantAndKratosUserID retrieves identities by (tenant_id, kratos_user_id).
func (r *userIdentityRepository) ListByTenantAndKratosUserID(
	ctx context.Context,
	tx *gorm.DB,
	tenantID, kratosUserID string,
) ([]*domain.UserIdentity, error) {
	db := r.db
	if tx != nil {
		db = tx
	}

	var identities []*domain.UserIdentity
	err := db.WithContext(ctx).
		Where(`
			tenant_id = ? 
			AND global_user_id = (
				SELECT global_user_id 
				FROM user_identities 
				WHERE tenant_id = ? AND kratos_user_id = ? 
				LIMIT 1
			)
		`, tenantID, tenantID, kratosUserID).
		Find(&identities).Error
	if err != nil {
		return nil, err
	}

	return identities, nil
}

// ExistsByTenantGlobalUserIDAndType checks by (tenant_id, global_user_id, type).
func (r *userIdentityRepository) ExistsByTenantGlobalUserIDAndType(
	ctx context.Context,
	tenantID, globalUserID, identityType string,
) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.UserIdentity{}).
		Where("tenant_id = ? AND global_user_id = ? AND type = ?", tenantID, globalUserID, identityType).
		Count(&count).Error
	return count > 0, err
}

// GetByGlobalUserID retrieves identities per tenant by (global_user_id).
func (r *userIdentityRepository) GetByGlobalUserIDAndTenantID(ctx context.Context, tx *gorm.DB, globalUserID, tenantID string) ([]*domain.UserIdentity, error) {
	db := r.db
	if tx != nil {
		db = tx
	}

	var identities []*domain.UserIdentity
	err := db.WithContext(ctx).
		Where("global_user_id = ? AND tenant_id = ?", globalUserID, tenantID).
		Find(&identities).Error
	if err != nil {
		return nil, err
	}
	return identities, nil
}

func (r *userIdentityRepository) Delete(tx *gorm.DB, identityID string) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Delete(&domain.UserIdentity{ID: identityID}).Error
}
