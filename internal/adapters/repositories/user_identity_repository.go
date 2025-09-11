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

// ListByTenantAndKratosUserIDWithLang returns all identities that share the same
// global_user_id as (tenant_id, kratos_user_id) and also returns the tenant-level lang.
// Runs in a single DB round-trip.
func (r *userIdentityRepository) ListByTenantAndKratosUserIDWithLang(
	ctx context.Context,
	tx *gorm.DB,
	tenantID, kratosUserID string,
) ([]*domain.UserIdentity, string, error) {
	db := r.db
	if tx != nil {
		db = tx
	}

	// Temporary row type to collect ui.* plus mapping.lang
	type row struct {
		domain.UserIdentity
		Lang *string `gorm:"column:lang"`
	}

	var rows []row
	err := db.WithContext(ctx).
		Table("user_identities ui").
		Select(`
			ui.id, ui.global_user_id, ui.type, ui.value, ui.created_at, ui.updated_at, ui.tenant_id, ui.kratos_user_id,
			m.lang
		`).
		Joins(`LEFT JOIN user_identifier_mapping m ON m.global_user_id = ui.global_user_id`).
		Where(`
			ui.tenant_id = ?
			AND ui.global_user_id = (
				SELECT global_user_id
				FROM user_identities
				WHERE tenant_id = ? AND kratos_user_id = ?
				LIMIT 1
			)
		`, tenantID, tenantID, kratosUserID).
		Order("ui.type ASC, ui.created_at ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, "", err
	}
	if len(rows) == 0 {
		return nil, "", gorm.ErrRecordNotFound
	}

	// Extract lang (same for the whole global_user_id), and build []*domain.UserIdentity
	lang := ""
	if rows[0].Lang != nil {
		lang = *rows[0].Lang
	}

	identities := make([]*domain.UserIdentity, 0, len(rows))
	for i := range rows {
		ri := rows[i].UserIdentity // copy to avoid &loopvar pitfall
		identities = append(identities, &ri)
	}

	return identities, lang, nil
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

func (r *userIdentityRepository) Delete(tx *gorm.DB, identityID string) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Delete(&domain.UserIdentity{ID: identityID}).Error
}
