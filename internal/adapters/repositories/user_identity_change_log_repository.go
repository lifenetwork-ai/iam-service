package repositories

import (
	"context"

	"gorm.io/gorm"

	interfaces "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

type userIdentityChangeLogRepository struct {
	db *gorm.DB
}

func NewUserIdentityChangeLogRepository(db *gorm.DB) interfaces.UserIdentityChangeLogRepository {
	return &userIdentityChangeLogRepository{db: db}
}

func (r *userIdentityChangeLogRepository) LogChange(tx *gorm.DB, log *domain.UserIdentityChangeLog) error {
	return tx.Create(log).Error
}

func (r *userIdentityChangeLogRepository) GetLogsByGlobalUserID(ctx context.Context, globalUserID string) ([]domain.UserIdentityChangeLog, error) {
	var logs []domain.UserIdentityChangeLog
	if err := r.db.WithContext(ctx).
		Where("global_user_id = ?", globalUserID).
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
