package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
	cachetypes "github.com/lifenetwork-ai/iam-service/infrastructures/caching/types"
	repotypes "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	entities "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

var (
	_ repotypes.TenantRepository = &CachedTenantRepository{}

	tenantCacheKey = "tenant:%v"
	tenantCacheTTL = 7 * 24 * time.Hour
)

// Cached Tenant Repository
type CachedTenantRepository struct {
	db    *gorm.DB
	cache cachetypes.CacheClient
}

func NewCachedTenantRepository(db *gorm.DB, cache cachetypes.CacheClient) *CachedTenantRepository {
	return &CachedTenantRepository{
		db:    db,
		cache: cache,
	}
}

func (r *CachedTenantRepository) getCacheKey(id interface{}) string {
	return fmt.Sprintf(tenantCacheKey, id)
}

func (r *CachedTenantRepository) Create(tenant *entities.Tenant) error {
	if tenant.ID == uuid.Nil {
		tenant.ID = uuid.New()
	}
	if err := r.db.Create(tenant).Error; err != nil {
		return err
	}

	// Cache the newly created tenant
	ctx := context.Background()
	key := r.getCacheKey(tenant.ID)
	if err := r.cache.Set(ctx, key, tenant, tenantCacheTTL); err != nil {
		logger.GetLogger().Errorf("Failed to cache tenant after creation: %v", err)
	}
	return nil
}

func (r *CachedTenantRepository) Update(tenant *entities.Tenant) error {
	if tenant.ID == uuid.Nil {
		return errors.New("tenant ID is required")
	}

	if err := r.db.Save(tenant).Error; err != nil {
		return err
	}

	// Update cache
	ctx := context.Background()
	key := r.getCacheKey(tenant.ID)
	if err := r.cache.Set(ctx, key, tenant, tenantCacheTTL); err != nil {
		logger.GetLogger().Errorf("Failed to update tenant in cache: %v", err)
	}
	return nil
}

func (r *CachedTenantRepository) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&entities.Tenant{}, id).Error; err != nil {
		return err
	}

	// Remove from cache
	ctx := context.Background()
	key := r.getCacheKey(id)
	if err := r.cache.Del(ctx, key); err != nil {
		logger.GetLogger().Errorf("Failed to remove tenant from cache: %v", err)
	}
	return nil
}

func (r *CachedTenantRepository) GetByName(name string) (*entities.Tenant, error) {
	// Try cache first
	ctx := context.Background()
	key := r.getCacheKey("name:" + name)
	var tenant entities.Tenant
	if err := r.cache.Get(ctx, key, &tenant); err == nil {
		return &tenant, nil
	}

	// Cache miss, query DB
	if err := r.db.Where("name = ?", name).First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// Cache the result
	if err := r.cache.Set(ctx, key, tenant, tenantCacheTTL); err != nil {
		logger.GetLogger().Errorf("Failed to cache tenant by name: %v", err)
	}

	// Also cache by ID for future GetByID calls
	idKey := r.getCacheKey(tenant.ID)
	if err := r.cache.Set(ctx, idKey, tenant, tenantCacheTTL); err != nil {
		logger.GetLogger().Errorf("Failed to cache tenant by ID: %v", err)
	}

	return &tenant, nil
}

func (r *CachedTenantRepository) GetByID(id uuid.UUID) (*entities.Tenant, error) {
	// Try cache first
	ctx := context.Background()
	key := r.getCacheKey(id)
	var tenant entities.Tenant
	if err := r.cache.Get(ctx, key, &tenant); err == nil {
		return &tenant, nil
	}

	// Cache miss, query DB
	if err := r.db.First(&tenant, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// Cache the result
	if err := r.cache.Set(ctx, key, tenant, tenantCacheTTL); err != nil {
		logger.GetLogger().Errorf("Failed to cache tenant: %v", err)
	}

	return &tenant, nil
}

func (r *CachedTenantRepository) List() ([]*entities.Tenant, error) {
	// For list operations, we'll use a shorter cache duration since this data changes more frequently
	ctx := context.Background()
	key := r.getCacheKey("all")
	var tenants []*entities.Tenant

	// Try cache first
	if err := r.cache.Get(ctx, key, &tenants); err == nil {
		return tenants, nil
	}

	// Cache miss, query DB
	if err := r.db.Find(&tenants).Error; err != nil {
		return nil, err
	}

	if err := r.cache.Set(ctx, key, tenants, tenantCacheTTL); err != nil {
		logger.GetLogger().Errorf("Failed to cache tenant list: %v", err)
	}

	return tenants, nil
}
