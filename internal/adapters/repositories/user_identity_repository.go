package repositories

import (
	"context"

	"gorm.io/gorm"

	interfaces "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

type userIdentityRepository struct {
	db *gorm.DB
}

func NewUserIdentityRepository(db *gorm.DB) interfaces.UserIdentityRepository {
	return &userIdentityRepository{db: db}
}

func (r *userIdentityRepository) GetByGlobalUserID(ctx context.Context, globalUserID string) ([]domain.UserIdentity, error) {
	var list []domain.UserIdentity
	if err := r.db.WithContext(ctx).Where("global_user_id = ?", globalUserID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *userIdentityRepository) GetByTypeAndValue(ctx context.Context, typ, value string) (*domain.UserIdentity, error) {
	var identity domain.UserIdentity
	if err := r.db.WithContext(ctx).
		Where("type = ? AND value = ?", typ, value).
		First(&identity).Error; err != nil {
		return nil, err
	}
	return &identity, nil
}

func (r *userIdentityRepository) FindGlobalUserIDByIdentity(ctx context.Context, typ, value string) (string, error) {
	var identity domain.UserIdentity
	if err := r.db.WithContext(ctx).
		Select("global_user_id").
		Where("type = ? AND value = ?", typ, value).
		First(&identity).Error; err != nil {
		return "", err
	}
	return identity.GlobalUserID, nil
}

func (r *userIdentityRepository) Create(tx *gorm.DB, identity *domain.UserIdentity) error {
	return tx.Create(identity).Error
}

func (r *userIdentityRepository) Update(tx *gorm.DB, identity *domain.UserIdentity) error {
	return tx.Save(identity).Error
}
