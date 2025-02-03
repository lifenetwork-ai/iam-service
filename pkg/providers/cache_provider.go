package providers

import (
	"context"
	"sync"

	"github.com/genefriendway/human-network-iam/infra/caching"
	"github.com/genefriendway/human-network-iam/infra/interfaces"
)

var (
	once      sync.Once
	cacheRepo interfaces.CacheRepository
)

// ProvideCacheRepository provides a singleton instance of CacheRepository.
func ProvideCacheRepository(ctx context.Context) interfaces.CacheRepository {
	once.Do(func() {
		cacheClient := caching.NewGoCacheClient()
		cacheRepo = caching.NewCachingRepository(ctx, cacheClient)
	})
	return cacheRepo
}
