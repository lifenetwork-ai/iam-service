package interfaces

import (
	"context"

	"gorm.io/gorm"

	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

type UserIdentifierMappingRepository interface {
	ExistsByTenantAndTenantUserID(
		ctx context.Context, tx *gorm.DB, tenant, tenantUserID string,
	) (bool, error)
	GetByGlobalUserID(ctx context.Context, globalUserID string) ([]domain.UserIdentifierMapping, error)
	ExistsMapping(ctx context.Context, tenant, globalUserID string) (bool, error)
	Create(tx *gorm.DB, mapping *domain.UserIdentifierMapping) error
}
