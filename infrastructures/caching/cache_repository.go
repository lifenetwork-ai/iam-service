package caching

import (
	"context"
	"fmt"
	"time"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/infrastructures/interfaces"
)

type cachingRepository struct {
	ctx    context.Context
	client interfaces.CacheClient
}

// NewCachingRepository initializes a new caching repository
func NewCachingRepository(ctx context.Context, client interfaces.CacheClient) interfaces.CacheRepository {
	return &cachingRepository{
		ctx:    ctx,
		client: client,
	}
}

// prependAppPrefix ensures all cache keys have a consistent prefix
func (repo *cachingRepository) prependAppPrefix(key string) string {
	return fmt.Sprintf("%s_%s", conf.GetAppName(), key)
}

// SaveItem saves an item to the cache with a specified expiration time
func (repo *cachingRepository) SaveItem(key fmt.Stringer, val interface{}, expire time.Duration) error {
	prefixedKey := repo.prependAppPrefix(key.String())
	return repo.client.Set(repo.ctx, prefixedKey, val, expire)
}

// RetrieveItem retrieves an item from the cache
func (repo *cachingRepository) RetrieveItem(key fmt.Stringer, val interface{}) error {
	prefixedKey := repo.prependAppPrefix(key.String())
	return repo.client.Get(repo.ctx, prefixedKey, val)
}

// RemoveItem removes an item from the cache
func (repo *cachingRepository) RemoveItem(key fmt.Stringer) error {
	prefixedKey := repo.prependAppPrefix(key.String())
	return repo.client.Del(repo.ctx, prefixedKey)
}
