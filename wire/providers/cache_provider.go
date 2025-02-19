package providers

import (
	"context"
	"sync"

	"github.com/genefriendway/human-network-iam/conf"
	"github.com/genefriendway/human-network-iam/infrastructures/caching"
	"github.com/genefriendway/human-network-iam/infrastructures/interfaces"
	"github.com/genefriendway/human-network-iam/packages/logger"
)

var (
	once      sync.Once
	cacheRepo interfaces.CacheRepository
)

// ProvideCacheRepository provides a singleton instance of CacheRepository.
func ProvideCacheRepository(ctx context.Context) interfaces.CacheRepository {
	once.Do(func() {
		cacheType := conf.GetCacheType()
		switch cacheType {
		case "redis":
			// Using Redis cache
			logger.GetLogger().Info("Using Redis cache")
			cacheClient := caching.NewRedisCacheClient()
			cacheRepo = caching.NewCachingRepository(ctx, cacheClient)
		default:
			// Using in-memory cache (default)
			logger.GetLogger().Info("Using in-memory cache (default)")
			cacheClient := caching.NewGoCacheClient()
			cacheRepo = caching.NewCachingRepository(ctx, cacheClient)
		}
	})
	return cacheRepo
}
