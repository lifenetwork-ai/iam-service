package repositories

import (
	"context"

	"gorm.io/gorm"

	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
)

type userIdentifierMappingRepository struct {
	db *gorm.DB
}

func NewUserIdentifierMappingRepository(db *gorm.DB) domainrepo.UserIdentifierMappingRepository {
	return &userIdentifierMappingRepository{db: db}
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

func (r *userIdentifierMappingRepository) ExistsMapping(
	ctx context.Context,
	tenantID string,
	globalUserID string,
) (bool, error) {
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

func (r *userIdentifierMappingRepository) Create(tx *gorm.DB, mapping *domain.UserIdentifierMapping) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(mapping).Error
}
