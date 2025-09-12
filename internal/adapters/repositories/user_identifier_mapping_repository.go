package repositories

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
)

type userIdentifierMappingRepository struct {
	db *gorm.DB
}

func NewUserIdentifierMappingRepository(db *gorm.DB) domainrepo.UserIdentifierMappingRepository {
	return &userIdentifierMappingRepository{db: db}
}

func (r *userIdentifierMappingRepository) GetByGlobalUserID(
	ctx context.Context,
	globalUserID string,
) (*domain.UserIdentifierMapping, error) {
	var mapping domain.UserIdentifierMapping
	if err := r.db.WithContext(ctx).
		Where("global_user_id = ?", globalUserID).
		First(&mapping).Error; err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (r *userIdentifierMappingRepository) ExistsByTenantAndKratosUserID(
	ctx context.Context, tx *gorm.DB, tenantID, kratosUserID string,
) (bool, error) {
	db := r.db
	if tx != nil {
		db = tx
	}

	var count int64
	err := db.WithContext(ctx).
		Table((&domain.UserIdentifierMapping{}).TableName()+" AS m").
		Joins("JOIN user_identities ui ON ui.global_user_id = m.global_user_id").
		Where("ui.tenant_id = ? AND ui.kratos_user_id = ?", tenantID, kratosUserID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userIdentifierMappingRepository) ExistsMapping(ctx context.Context, tenantID, globalUserID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.UserIdentifierMapping{}).
		Joins("JOIN user_identities ui ON ui.global_user_id = user_identifier_mapping.global_user_id").
		Where("ui.tenant_id = ? AND user_identifier_mapping.global_user_id = ?", tenantID, globalUserID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userIdentifierMappingRepository) GetByTenantIDAndKratosUserID(
	ctx context.Context, tenantID, kratosUserID string,
) (*domain.UserIdentifierMapping, error) {
	var mapping domain.UserIdentifierMapping
	err := r.db.WithContext(ctx).
		Model(&domain.UserIdentifierMapping{}).
		Joins("JOIN user_identities ui ON ui.global_user_id = user_identifier_mapping.global_user_id").
		Where("ui.tenant_id = ? AND ui.kratos_user_id = ?", tenantID, kratosUserID).
		First(&mapping).Error
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (r *userIdentifierMappingRepository) GetByTenantIDAndTenantUserID(
	ctx context.Context, tenantID, tenantUserID string,
) (*domain.UserIdentifierMapping, error) {
	var mapping domain.UserIdentifierMapping
	err := r.db.WithContext(ctx).
		Model(&domain.UserIdentifierMapping{}).
		Where("tenant_id = ? AND tenant_user_id = ?", tenantID, tenantUserID).
		First(&mapping).Error
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (r *userIdentifierMappingRepository) GetByGlobalUserIDAndTenantID(ctx context.Context, globalUserID, tenantID string) ([]*domain.UserIdentifierMapping, error) {
	var mapping []*domain.UserIdentifierMapping
	if err := r.db.WithContext(ctx).
		Where("global_user_id = ? AND tenant_id = ?", globalUserID, tenantID).
		Find(&mapping).Error; err != nil {
		return nil, err
	}
	return mapping, nil
}

func (r *userIdentifierMappingRepository) GetByTypeAndValue(
	ctx context.Context, tenantID, identifierType, identifierValue string,
) (string, error) {
	var mapping domain.UserIdentifierMapping
	if err := r.db.WithContext(ctx).
		Joins("JOIN user_identities ui ON ui.global_user_id = user_identifier_mapping.global_user_id").
		Where("ui.tenant_id = ? AND ui.type = ? AND ui.value = ?", tenantID, identifierType, identifierValue).
		First(&mapping).Error; err != nil {
		return "", err
	}
	return mapping.GlobalUserID, nil
}

func (r *userIdentifierMappingRepository) Create(ctx context.Context, tx *gorm.DB, mapping *domain.UserIdentifierMapping) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Create(mapping).Error
}

// Upsert creates or updates the entire mapping row by global_user_id.
func (r *userIdentifierMappingRepository) Upsert(
	ctx context.Context,
	tx *gorm.DB,
	mapping *domain.UserIdentifierMapping,
) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "global_user_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"lang": mapping.Lang,
			}),
		}).
		Create(mapping).Error
}

func (r *userIdentifierMappingRepository) Update(tx *gorm.DB, mapping *domain.UserIdentifierMapping) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Model(&domain.UserIdentifierMapping{}).Where("id = ?", mapping.ID).Updates(mapping).Error
}
