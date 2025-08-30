package domain

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

type AdminAccountRepository interface {
	Create(account *domain.AdminAccount) error
	GetByUsername(username string) (*domain.AdminAccount, error)
	GetByID(id string) (*domain.AdminAccount, error)
	Update(account *domain.AdminAccount) error
	Delete(id string) error
}

type ChallengeSessionRepository interface {
	SaveChallenge(ctx context.Context, sessionID string, challenge *domain.ChallengeSession, ttl time.Duration) error
	GetChallenge(ctx context.Context, sessionID string) (*domain.ChallengeSession, error)
	DeleteChallenge(ctx context.Context, sessionID string) error
}

type GlobalUserRepository interface {
	GetByID(ctx context.Context, id string) (*domain.GlobalUser, error)
	Create(tx *gorm.DB, user *domain.GlobalUser) error
}

type TenantRepository interface {
	Create(tenant *domain.Tenant) error
	Update(tenant *domain.Tenant) error
	Delete(id uuid.UUID) error
	GetByID(id uuid.UUID) (*domain.Tenant, error)
	List() ([]*domain.Tenant, error)
	GetByName(name string) (*domain.Tenant, error)
}

type UserIdentityChangeLogRepository interface {
	Create(ctx context.Context, tx *gorm.DB, log *domain.UserIdentityChangeLog) error
	ListByGlobalUserID(ctx context.Context, globalUserID uuid.UUID) ([]*domain.UserIdentityChangeLog, error)
	ListByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*domain.UserIdentityChangeLog, error)
}

type UserIdentifierMappingRepository interface {
	ExistsByTenantAndTenantUserID(ctx context.Context, tx *gorm.DB, tenantID, tenantUserID string) (bool, error)
	ExistsMapping(ctx context.Context, tenantID, globalUserID string) (bool, error)
	Create(tx *gorm.DB, mapping *domain.UserIdentifierMapping) error
	GetByTenantIDAndIdentifier(ctx context.Context, tenantID, identifierType, identifierValue string) (string, error)
	GetByTenantIDAndTenantUserID(ctx context.Context, tenantID, tenantUserID string) (*domain.UserIdentifierMapping, error)
	Update(tx *gorm.DB, mapping *domain.UserIdentifierMapping) error
}

type UserIdentityRepository interface {
	GetByTypeAndValue(ctx context.Context, tx *gorm.DB, identityType, value string) (*domain.UserIdentity, error)
	FindGlobalUserIDByIdentity(ctx context.Context, tenantID, identityType, value string) (string, error)
	InsertOnceByUserAndType(ctx context.Context, globalUserID, idType, value string) (bool, error)
	FirstOrCreate(tx *gorm.DB, identity *domain.UserIdentity) error
	Update(tx *gorm.DB, identity *domain.UserIdentity) error
	ExistsWithinTenant(ctx context.Context, tenantID, identityType, value string) (bool, error)
	GetByTenantAndTenantUserID(ctx context.Context, tx *gorm.DB, tenantID, tenantUserID string) (*domain.UserIdentity, error)
	ExistsByGlobalUserIDAndType(ctx context.Context, globalUserID, identityType string) (bool, error)
	Delete(tx *gorm.DB, identityID string) error
}
