package caching

import (
	"context"
	"reflect"
	"time"

	"github.com/lifenetwork-ai/iam-service/infrastructures/caching/types"
	"github.com/patrickmn/go-cache"
)

type goCacheClient struct {
	cache *cache.Cache
}

// NewGoCacheClient initializes a new cache client with default expiration and cleanup interval
func NewGoCacheClient(client *cache.Cache) types.CacheClient {
	return &goCacheClient{
		cache: client,
	}
}

// Set adds an item to the cache with a specified expiration time
// If the duration is 0 (DefaultExpiration), the cache's default expiration time is used.
// If it is -1 (NoExpiration), the item never expires.
func (c *goCacheClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	c.cache.Set(key, value, expiration)
	return nil
}

func (c *goCacheClient) Get(ctx context.Context, key string, dest interface{}) error {
	cachedValue, found := c.cache.Get(key)
	if !found {
		return types.ErrCacheMiss
	}

	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr || destVal.IsNil() {
		return types.ErrInvalidDestination
	}

	cachedVal := reflect.ValueOf(cachedValue)
	destType := destVal.Elem().Type()

	// Case 1: Direct assignment (same types)
	if cachedVal.Type().AssignableTo(destType) {
		destVal.Elem().Set(cachedVal)
		return nil
	}

	// Case 2: Cached is pointer, dest is value (*T -> T)
	if cachedVal.Kind() == reflect.Ptr && !cachedVal.IsNil() && cachedVal.Elem().Type().AssignableTo(destType) {
		destVal.Elem().Set(cachedVal.Elem())
		return nil
	}

	// Case 3: Cached is value, dest is pointer (T -> *T)
	if destType.Kind() == reflect.Ptr && cachedVal.Type().AssignableTo(destType.Elem()) {
		newPtr := reflect.New(destType.Elem())
		newPtr.Elem().Set(cachedVal)
		destVal.Elem().Set(newPtr)
		return nil
	}

	// Case 4: Both are pointers but different levels (*T -> **T or **T -> *T)
	if cachedVal.Kind() == reflect.Ptr && destType.Kind() == reflect.Ptr {
		if !cachedVal.IsNil() && cachedVal.Elem().Type().AssignableTo(destType.Elem()) {
			destVal.Elem().Set(cachedVal)
			return nil
		}
	}

	return types.ErrTypeMismatch
}

// Del deletes an item from the cache
func (c *goCacheClient) Del(ctx context.Context, key string) error {
	c.cache.Delete(key)
	return nil
}
