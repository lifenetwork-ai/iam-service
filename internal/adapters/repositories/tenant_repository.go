package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	infra_interfaces "github.com/lifenetwork-ai/iam-service/infrastructures/interfaces"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
	"gorm.io/gorm"

	repo_types "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
)

var _ repo_types.TenantRepository = &TenantRepository{}
var _ repo_types.TenantRepository = &CachedTenantRepository{}

var (
	tenantCacheKey = "tenant:%v"
	tenantCacheTTL = 7 * 24 * time.Hour
)

type TenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{
		db: db,
	}
}

func (r *TenantRepository) Create(tenant *repo_types.Tenant) error {
	if tenant.ID == uuid.Nil {
		tenant.ID = uuid.New()
	}
	return r.db.Create(tenant).Error
}

func (r *TenantRepository) Update(tenant *repo_types.Tenant) error {
	if tenant.ID == uuid.Nil {
		return errors.New("tenant ID is required")
	}
	return r.db.Save(tenant).Error
}

func (r *TenantRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&repo_types.Tenant{}, id).Error
}

func (r *TenantRepository) GetByID(id uuid.UUID) (*repo_types.Tenant, error) {
	var tenant repo_types.Tenant
	if err := r.db.First(&tenant, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepository) List() ([]*repo_types.Tenant, error) {
	var tenants []*repo_types.Tenant
	if err := r.db.Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}

func (r *TenantRepository) GetByName(name string) (*repo_types.Tenant, error) {
	var tenant repo_types.Tenant
	if err := r.db.Where("name = ?", name).First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tenant, nil
}

// Cached Tenant Repository

type CachedTenantRepository struct {
	db    *gorm.DB
	cache infra_interfaces.CacheClient
}

func NewCachedTenantRepository(db *gorm.DB, cache infra_interfaces.CacheClient) *CachedTenantRepository {
	return &CachedTenantRepository{
		db:    db,
		cache: cache,
	}
}

func (r *CachedTenantRepository) getCacheKey(id interface{}) string {
	return fmt.Sprintf(tenantCacheKey, id)
}

func (r *CachedTenantRepository) Create(tenant *repo_types.Tenant) error {
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

func (r *CachedTenantRepository) Update(tenant *repo_types.Tenant) error {
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
	if err := r.db.Delete(&repo_types.Tenant{}, id).Error; err != nil {
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

func (r *CachedTenantRepository) GetByName(name string) (*repo_types.Tenant, error) {
	// Try cache first
	ctx := context.Background()
	key := r.getCacheKey("name:" + name)
	var tenant repo_types.Tenant
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

func (r *CachedTenantRepository) GetByID(id uuid.UUID) (*repo_types.Tenant, error) {
	// Try cache first
	ctx := context.Background()
	key := r.getCacheKey(id)
	var tenant repo_types.Tenant
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

func (r *CachedTenantRepository) List() ([]*repo_types.Tenant, error) {
	// For list operations, we'll use a shorter cache duration since this data changes more frequently
	ctx := context.Background()
	key := r.getCacheKey("all")
	var tenants []*repo_types.Tenant

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
