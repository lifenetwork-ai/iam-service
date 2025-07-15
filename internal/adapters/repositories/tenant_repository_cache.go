package repositories

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
	cachetypes "github.com/lifenetwork-ai/iam-service/infrastructures/caching/types"
	repotypes "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	entities "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

const tenantCacheTTL = 7 * 24 * time.Hour

var _ repotypes.TenantRepository = &tenantRepositoryCache{}

type tenantRepositoryCache struct {
	db    *gorm.DB
	cache cachetypes.CacheRepository
}

func NewCachedTenantRepository(db *gorm.DB, cache cachetypes.CacheRepository) *tenantRepositoryCache {
	return &tenantRepositoryCache{
		db:    db,
		cache: cache,
	}
}

func getTenantCacheKey(key interface{}) *cachetypes.Keyer {
	return &cachetypes.Keyer{Raw: fmt.Sprintf("tenant:%v", key)}
}

func (r *tenantRepositoryCache) Create(tenant *entities.Tenant) error {
	if tenant.ID == uuid.Nil {
		tenant.ID = uuid.New()
	}

	if err := r.db.Create(tenant).Error; err != nil {
		return err
	}

	// Cache by ID and name
	_ = r.cache.SaveItem(getTenantCacheKey(tenant.ID), tenant, tenantCacheTTL)
	_ = r.cache.SaveItem(getTenantCacheKey("name:"+tenant.Name), tenant, tenantCacheTTL)
	return nil
}

func (r *tenantRepositoryCache) Update(tenant *entities.Tenant) error {
	if tenant.ID == uuid.Nil {
		return errors.New("tenant ID is required")
	}

	if err := r.db.Save(tenant).Error; err != nil {
		return err
	}

	_ = r.cache.SaveItem(getTenantCacheKey(tenant.ID), tenant, tenantCacheTTL)
	_ = r.cache.SaveItem(getTenantCacheKey("name:"+tenant.Name), tenant, tenantCacheTTL)
	return nil
}

func (r *tenantRepositoryCache) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&entities.Tenant{}, id).Error; err != nil {
		return err
	}

	_ = r.cache.RemoveItem(getTenantCacheKey(id))
	return nil
}

func (r *tenantRepositoryCache) GetByID(id uuid.UUID) (*entities.Tenant, error) {
	var tenant entities.Tenant
	if err := r.cache.RetrieveItem(getTenantCacheKey(id), &tenant); err == nil {
		return &tenant, nil
	}

	if err := r.db.First(&tenant, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	_ = r.cache.SaveItem(getTenantCacheKey(id), tenant, tenantCacheTTL)
	_ = r.cache.SaveItem(getTenantCacheKey("name:"+tenant.Name), tenant, tenantCacheTTL)

	return &tenant, nil
}

func (r *tenantRepositoryCache) GetByName(name string) (*entities.Tenant, error) {
	var tenant entities.Tenant
	if err := r.cache.RetrieveItem(getTenantCacheKey("name:"+name), &tenant); err == nil {
		return &tenant, nil
	}

	if err := r.db.Where("name = ?", name).First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	_ = r.cache.SaveItem(getTenantCacheKey("name:"+name), tenant, tenantCacheTTL)
	_ = r.cache.SaveItem(getTenantCacheKey(tenant.ID), tenant, tenantCacheTTL)

	return &tenant, nil
}

func (r *tenantRepositoryCache) List() ([]*entities.Tenant, error) {
	var tenants []*entities.Tenant
	key := getTenantCacheKey("all")

	if err := r.cache.RetrieveItem(key, &tenants); err == nil {
		return tenants, nil
	}

	if err := r.db.Find(&tenants).Error; err != nil {
		return nil, err
	}

	_ = r.cache.SaveItem(key, tenants, tenantCacheTTL)
	return tenants, nil
}
