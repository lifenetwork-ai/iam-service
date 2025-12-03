package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	cachetypes "github.com/lifenetwork-ai/iam-service/infrastructures/caching/types"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

const zaloTokenCacheTTL = 1 * time.Hour

type zaloTokenRepositoryCache struct {
	repo  domainrepo.ZaloTokenRepository
	cache cachetypes.CacheRepository
}

func NewZaloTokenRepositoryCache(
	repo domainrepo.ZaloTokenRepository,
	cache cachetypes.CacheRepository,
) domainrepo.ZaloTokenRepository {
	return &zaloTokenRepositoryCache{
		repo:  repo,
		cache: cache,
	}
}

func zaloTokenKey(tenantID uuid.UUID) *cachetypes.Keyer {
	return &cachetypes.Keyer{Raw: fmt.Sprintf("zalo:token:%s", tenantID.String())}
}

// Get retrieves the Zalo token for a specific tenant
func (c *zaloTokenRepositoryCache) Get(ctx context.Context, tenantID uuid.UUID) (*domain.ZaloToken, error) {
	var token domain.ZaloToken
	if err := c.cache.RetrieveItem(zaloTokenKey(tenantID), &token); err == nil {
		return &token, nil
	}

	tokenPtr, err := c.repo.Get(ctx, tenantID)
	if err != nil || tokenPtr == nil {
		return tokenPtr, err
	}

	err = c.cache.SaveItem(zaloTokenKey(tenantID), tokenPtr, zaloTokenCacheTTL)
	if err != nil {
		logger.GetLogger().Errorf("Failed to save zalo token to cache: %v", err)
	}

	return tokenPtr, nil
}

// Save creates or updates the token for a tenant
func (c *zaloTokenRepositoryCache) Save(ctx context.Context, token *domain.ZaloToken) error {
	if err := c.repo.Save(ctx, token); err != nil {
		return err
	}

	err := c.cache.SaveItem(zaloTokenKey(token.TenantID), token, zaloTokenCacheTTL)
	if err != nil {
		logger.GetLogger().Errorf("Failed to save zalo token to cache: %v", err)
	}

	return nil
}

// GetAll retrieves all tokens (for the refresh worker) - no caching
func (c *zaloTokenRepositoryCache) GetAll(ctx context.Context) ([]*domain.ZaloToken, error) {
	return c.repo.GetAll(ctx)
}

// Delete removes a tenant's Zalo token configuration
func (c *zaloTokenRepositoryCache) Delete(ctx context.Context, tenantID uuid.UUID) error {
	if err := c.repo.Delete(ctx, tenantID); err != nil {
		return err
	}

	// Invalidate cache
	err := c.cache.RemoveItem(zaloTokenKey(tenantID))
	if err != nil {
		logger.GetLogger().Errorf("Failed to delete zalo token from cache: %v", err)
	}

	return nil
}
