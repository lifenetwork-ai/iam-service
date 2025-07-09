package repositories

import (
	"context"

	"gorm.io/gorm"

	interfaces "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

type userIdentifierMappingRepository struct {
	db *gorm.DB
}

func NewUserIdentifierMappingRepository(db *gorm.DB) interfaces.UserIdentifierMappingRepository {
	return &userIdentifierMappingRepository{db: db}
}

func (r *userIdentifierMappingRepository) ExistsByTenantAndTenantUserID(
	ctx context.Context, tx *gorm.DB, tenant, tenantUserID string,
) (bool, error) {
	var count int64
	if err := tx.WithContext(ctx).
		Model(&domain.UserIdentifierMapping{}).
		Where("tenant = ? AND tenant_user_id = ?", tenant, tenantUserID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userIdentifierMappingRepository) GetByGlobalUserID(ctx context.Context, globalUserID string) ([]domain.UserIdentifierMapping, error) {
	var mappings []domain.UserIdentifierMapping
	if err := r.db.WithContext(ctx).
		Where("global_user_id = ?", globalUserID).
		Find(&mappings).Error; err != nil {
		return nil, err
	}
	return mappings, nil
}

func (r *userIdentifierMappingRepository) ExistsMapping(ctx context.Context, tenant, globalUserID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.UserIdentifierMapping{}).
		Where("tenant = ? AND global_user_id = ?", tenant, globalUserID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userIdentifierMappingRepository) Create(tx *gorm.DB, mapping *domain.UserIdentifierMapping) error {
	return tx.Create(mapping).Error
}
