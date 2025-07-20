package interfaces

import (
	"context"

	"gorm.io/gorm"

	"github.com/google/uuid"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

type UserIdentityChangeLogRepository interface {
	Create(ctx context.Context, tx *gorm.DB, log *domain.UserIdentityChangeLog) error
	ListByGlobalUserID(ctx context.Context, globalUserID uuid.UUID) ([]*domain.UserIdentityChangeLog, error)
	ListByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*domain.UserIdentityChangeLog, error)
}
