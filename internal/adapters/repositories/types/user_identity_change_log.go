package interfaces

import (
	"context"

	"gorm.io/gorm"

	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

type UserIdentityChangeLogRepository interface {
	LogChange(tx *gorm.DB, log *domain.UserIdentityChangeLog) error
	GetLogsByGlobalUserID(ctx context.Context, globalUserID string) ([]domain.UserIdentityChangeLog, error)
}
