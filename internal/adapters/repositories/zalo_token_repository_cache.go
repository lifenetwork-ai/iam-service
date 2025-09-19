package repositories

import (
	"context"
	"time"

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

func zaloTokenKey() *cachetypes.Keyer {
	return &cachetypes.Keyer{Raw: "zalo:token"}
}

func (c *zaloTokenRepositoryCache) Get(ctx context.Context) (*domain.ZaloToken, error) {
	var token domain.ZaloToken
	if err := c.cache.RetrieveItem(zaloTokenKey(), &token); err == nil {
		return &token, nil
	}

	tokenPtr, err := c.repo.Get(ctx)
	if err != nil || tokenPtr == nil {
		return tokenPtr, err
	}

	err = c.cache.SaveItem(zaloTokenKey(), tokenPtr, zaloTokenCacheTTL)
	if err != nil {
		logger.GetLogger().Errorf("Failed to save zalo token to cache: %v", err)
	}

	return tokenPtr, nil
}

func (c *zaloTokenRepositoryCache) Save(ctx context.Context, token *domain.ZaloToken) error {
	if err := c.repo.Save(ctx, token); err != nil {
		return err
	}

	err := c.cache.SaveItem(zaloTokenKey(), token, zaloTokenCacheTTL)
	if err != nil {
		logger.GetLogger().Errorf("Failed to save zalo token to cache: %v", err)
	}

	return nil
}
