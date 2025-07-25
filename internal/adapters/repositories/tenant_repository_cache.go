package repositories

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	cachetypes "github.com/lifenetwork-ai/iam-service/infrastructures/caching/types"
	entities "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

const tenantCacheTTL = 7 * 24 * time.Hour

type tenantRepositoryCache struct {
	repo  domainrepo.TenantRepository
	cache cachetypes.CacheRepository
}

func NewTenantRepositoryCache(
	repo domainrepo.TenantRepository,
	cache cachetypes.CacheRepository,
) domainrepo.TenantRepository {
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
	err := c.cache.SaveItem(tenantKeyByID(tenant.ID), tenant, tenantCacheTTL)
	if err != nil {
		logger.GetLogger().Errorf("Failed to save tenant to cache: %v", err)
	}
	err = c.cache.SaveItem(tenantKeyByName(tenant.Name), tenant, tenantCacheTTL)
	if err != nil {
		logger.GetLogger().Errorf("Failed to save tenant to cache: %v", err)
	}

	err = c.updateListCache()
	if err != nil {
		logger.GetLogger().Errorf("Failed to update list cache: %v", err)
	}
	return nil
}

func (c *tenantRepositoryCache) Update(tenant *entities.Tenant) error {
	if err := c.repo.Update(tenant); err != nil {
		return err
	}
	err := c.cache.SaveItem(tenantKeyByID(tenant.ID), tenant, tenantCacheTTL)
	if err != nil {
		logger.GetLogger().Errorf("Failed to save tenant to cache: %v", err)
	}
	err = c.cache.SaveItem(tenantKeyByName(tenant.Name), tenant, tenantCacheTTL)
	if err != nil {
		logger.GetLogger().Errorf("Failed to save tenant to cache: %v", err)
	}

	err = c.updateListCache()
	if err != nil {
		logger.GetLogger().Errorf("Failed to update list cache: %v", err)
	}

	return nil
}

func (c *tenantRepositoryCache) Delete(id uuid.UUID) error {
	tenant, _ := c.repo.GetByID(id)
	if err := c.repo.Delete(id); err != nil {
		return err
	}
	err := c.cache.RemoveItem(tenantKeyByID(id))
	if err != nil {
		logger.GetLogger().Errorf("Failed to remove tenant from cache: %v", err)
	}
	if tenant != nil {
		err = c.cache.RemoveItem(tenantKeyByName(tenant.Name))
		if err != nil {
			logger.GetLogger().Errorf("Failed to remove tenant from cache: %v", err)
		}
	}

	// Update list cache
	err = c.updateListCache()
	if err != nil {
		logger.GetLogger().Errorf("Failed to update list cache: %v", err)
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

	err = c.cache.SaveItem(tenantKeyByID(tenantPtr.ID), tenantPtr, tenantCacheTTL)
	if err != nil {
		logger.GetLogger().Errorf("Failed to save tenant to cache: %v", err)
	}

	err = c.cache.SaveItem(tenantKeyByName(tenantPtr.Name), tenantPtr, tenantCacheTTL)
	if err != nil {
		logger.GetLogger().Errorf("Failed to save tenant to cache: %v", err)
	}

	// Update list cache
	err = c.updateListCache()
	if err != nil {
		logger.GetLogger().Errorf("Failed to update list cache: %v", err)
	}
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

	err = c.cache.SaveItem(tenantKeyByName(tenantPtr.Name), tenantPtr, tenantCacheTTL)
	if err != nil {
		logger.GetLogger().Errorf("Failed to save tenant to cache: %v", err)
	}

	err = c.cache.SaveItem(tenantKeyByID(tenantPtr.ID), tenantPtr, tenantCacheTTL)
	if err != nil {
		logger.GetLogger().Errorf("Failed to save tenant to cache: %v", err)
	}

	// Update list cache
	err = c.updateListCache()
	if err != nil {
		logger.GetLogger().Errorf("Failed to update list cache: %v", err)
	}
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

	err = c.cache.SaveItem(tenantKeyAll(), tenants, tenantCacheTTL)
	if err != nil {
		logger.GetLogger().Errorf("Failed to save tenant to cache: %v", err)
	}

	return tenants, nil
}

func (c *tenantRepositoryCache) updateListCache() error {
	tenants, err := c.repo.List()
	if err != nil {
		return err
	}
	return c.cache.SaveItem(tenantKeyAll(), tenants, tenantCacheTTL)
}
