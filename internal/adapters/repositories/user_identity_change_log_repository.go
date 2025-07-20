package repositories

import (
	"context"

	"gorm.io/gorm"

	"github.com/google/uuid"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

type userIdentityChangeLogRepository struct {
	db *gorm.DB
}

func NewUserIdentityChangeLogRepository(db *gorm.DB) interfaces.UserIdentityChangeLogRepository {
	return &userIdentityChangeLogRepository{db: db}
}

func (r *userIdentityChangeLogRepository) Create(ctx context.Context, tx *gorm.DB, log *domain.UserIdentityChangeLog) error {
	db := r.db.WithContext(ctx)
	if tx != nil {
		db = tx.WithContext(ctx)
	}
	return db.Create(log).Error
}

func (r *userIdentityChangeLogRepository) ListByGlobalUserID(ctx context.Context, globalUserID uuid.UUID) ([]*domain.UserIdentityChangeLog, error) {
	var logs []*domain.UserIdentityChangeLog
	err := r.db.WithContext(ctx).
		Where("global_user_id = ?", globalUserID).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

func (r *userIdentityChangeLogRepository) ListByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*domain.UserIdentityChangeLog, error) {
	var logs []*domain.UserIdentityChangeLog
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}
