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

func (r *userIdentifierMappingRepository) ExistsByTenantAndTenantUserID(
	ctx context.Context, tx *gorm.DB, tenantID, tenantUserID string,
) (bool, error) {
	var count int64
	if err := tx.WithContext(ctx).
		Model(&domain.UserIdentifierMapping{}).
		Where("tenant_id = ? AND tenant_user_id = ?", tenantID, tenantUserID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userIdentifierMappingRepository) GetByTenantIDAndTenantUserID(ctx context.Context, tenantID, tenantUserID string) (*domain.UserIdentifierMapping, error) {
	var mapping domain.UserIdentifierMapping
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND tenant_user_id = ?", tenantID, tenantUserID).
		First(&mapping).Error; err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (r *userIdentifierMappingRepository) GetByTenantIDAndIdentifier(ctx context.Context, tenantID, identifierType, identifierValue string) (string, error) {
	var mapping domain.UserIdentifierMapping
	if err := r.db.WithContext(ctx).
		Joins("JOIN user_identities ON user_identities.global_user_id = user_identifier_mapping.global_user_id").
		Where("user_identities.type = ? AND user_identities.value = ? AND user_identifier_mapping.tenant_id = ?", identifierType, identifierValue, tenantID).
		First(&mapping).Error; err != nil {
		return "", err
	}
	return mapping.TenantUserID, nil
}

func (r *userIdentifierMappingRepository) GetByGlobalUserIDAndTenantID(ctx context.Context, globalUserID, tenantID string) ([]*domain.UserIdentifierMapping, error) {
	var mapping []*domain.UserIdentifierMapping
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND global_user_id = ?", tenantID, globalUserID).
		Find(&mapping).Error; err != nil {
		return nil, err
	}
	return mapping, nil
}

func (r *userIdentifierMappingRepository) ExistsMapping(ctx context.Context, tenantID, globalUserID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.UserIdentifierMapping{}).
		Where("tenant_id = ? AND global_user_id = ?", tenantID, globalUserID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userIdentifierMappingRepository) Create(tx *gorm.DB, mapping *domain.UserIdentifierMapping) error {
	return tx.Create(mapping).Error
}

func (r *userIdentifierMappingRepository) Update(tx *gorm.DB, mapping *domain.UserIdentifierMapping) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Model(&domain.UserIdentifierMapping{}).Where("id = ?", mapping.ID).Updates(mapping).Error
}
