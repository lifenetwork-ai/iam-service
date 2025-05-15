package instances

import (
	"context"
	"sync"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/infrastructures/caching"
	"github.com/lifenetwork-ai/iam-service/infrastructures/interfaces"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

var (
	cacheOnce sync.Once
	cacheRepo interfaces.CacheRepository
)

// CacheRepositoryInstance provides a singleton instance of CacheRepository.
func CacheRepositoryInstance(ctx context.Context) interfaces.CacheRepository {
	cacheOnce.Do(func() {
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
