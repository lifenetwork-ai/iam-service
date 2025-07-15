package repositories

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	cachetypes "github.com/lifenetwork-ai/iam-service/infrastructures/caching/types"
	repotypes "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	entities "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

const tenantCacheTTL = 7 * 24 * time.Hour

type tenantRepositoryCache struct {
	repo  repotypes.TenantRepository
	cache cachetypes.CacheRepository
}

func NewCTenantRepositoryCache(
	repo repotypes.TenantRepository,
	cache cachetypes.CacheRepository,
) repotypes.TenantRepository {
	return &tenantRepositoryCache{
		repo:  repo,
		cache: cache,
	}
}

func tenantKeyByID(id uuid.UUID) *cachetypes.Keyer {
	return &cachetypes.Keyer{Raw: fmt.Sprintf("tenant:%s", id.String())}
}

func tenantKeyByName(name string) *cachetypes.Keyer {
	return &cachetypes.Keyer{Raw: fmt.Sprintf("tenant:name:%s", name)}
}

func tenantKeyAll() *cachetypes.Keyer {
	return &cachetypes.Keyer{Raw: "tenant:all"}
}

func (c *tenantRepositoryCache) Create(tenant *entities.Tenant) error {
	if err := c.repo.Create(tenant); err != nil {
		return err
	}
	// Cache both by ID and name
	_ = c.cache.SaveItem(tenantKeyByID(tenant.ID), tenant, tenantCacheTTL)
	_ = c.cache.SaveItem(tenantKeyByName(tenant.Name), tenant, tenantCacheTTL)
	return nil
}

func (c *tenantRepositoryCache) Update(tenant *entities.Tenant) error {
	if err := c.repo.Update(tenant); err != nil {
		return err
	}
	_ = c.cache.SaveItem(tenantKeyByID(tenant.ID), tenant, tenantCacheTTL)
	_ = c.cache.SaveItem(tenantKeyByName(tenant.Name), tenant, tenantCacheTTL)
	return nil
}

func (c *tenantRepositoryCache) Delete(id uuid.UUID) error {
	tenant, _ := c.repo.GetByID(id)
	if err := c.repo.Delete(id); err != nil {
		return err
	}
	_ = c.cache.RemoveItem(tenantKeyByID(id))
	if tenant != nil {
		_ = c.cache.RemoveItem(tenantKeyByName(tenant.Name))
	}
	return nil
}

func (c *tenantRepositoryCache) GetByID(id uuid.UUID) (*entities.Tenant, error) {
	var tenant entities.Tenant
	if err := c.cache.RetrieveItem(tenantKeyByID(id), &tenant); err == nil {
		return &tenant, nil
	}

	tenantPtr, err := c.repo.GetByID(id)
	if err != nil || tenantPtr == nil {
		return tenantPtr, err
	}

	_ = c.cache.SaveItem(tenantKeyByID(tenantPtr.ID), tenantPtr, tenantCacheTTL)
	_ = c.cache.SaveItem(tenantKeyByName(tenantPtr.Name), tenantPtr, tenantCacheTTL)
	return tenantPtr, nil
}

func (c *tenantRepositoryCache) GetByName(name string) (*entities.Tenant, error) {
	var tenant entities.Tenant
	if err := c.cache.RetrieveItem(tenantKeyByName(name), &tenant); err == nil {
		return &tenant, nil
	}

	tenantPtr, err := c.repo.GetByName(name)
	if err != nil || tenantPtr == nil {
		return tenantPtr, err
	}

	_ = c.cache.SaveItem(tenantKeyByName(tenantPtr.Name), tenantPtr, tenantCacheTTL)
	_ = c.cache.SaveItem(tenantKeyByID(tenantPtr.ID), tenantPtr, tenantCacheTTL)
	return tenantPtr, nil
}

func (c *tenantRepositoryCache) List() ([]*entities.Tenant, error) {
	var tenants []*entities.Tenant
	if err := c.cache.RetrieveItem(tenantKeyAll(), &tenants); err == nil {
		return tenants, nil
	}

	tenants, err := c.repo.List()
	if err != nil {
		return nil, err
	}

	_ = c.cache.SaveItem(tenantKeyAll(), tenants, tenantCacheTTL)
	return tenants, nil
}
