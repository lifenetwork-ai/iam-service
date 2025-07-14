package interfaces

import (
	"context"

	"gorm.io/gorm"

	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

type UserIdentityRepository interface {
	GetByGlobalUserID(ctx context.Context, globalUserID string) ([]domain.UserIdentity, error)
	GetByTypeAndValue(ctx context.Context, tx *gorm.DB, identityType, value string) (*domain.UserIdentity, error)
	FindGlobalUserIDByIdentity(ctx context.Context, identityType, value string) (string, error)
	FirstOrCreate(tx *gorm.DB, identity *domain.UserIdentity) error
	Update(tx *gorm.DB, identity *domain.UserIdentity) error
	ExistsWithinTenant(ctx context.Context, tenantID, identityType, value string) (bool, error)
}
