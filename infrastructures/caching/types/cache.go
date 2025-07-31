package types

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrCacheMiss          = errors.New("cache: item not found")
	ErrInvalidDestination = errors.New("cache: destination must be a non-nil pointer")
	ErrTypeMismatch       = errors.New("cache: type mismatch between cached value and destination")
)

type Keyer struct {
	Raw string
}

type Value struct {
	Raw string
}

func (k *Keyer) String() string {
	return k.Raw
}

type CacheClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Del(ctx context.Context, key string) error
}

type CacheRepository interface {
	SaveItem(key fmt.Stringer, val interface{}, expire time.Duration) error
	RetrieveItem(key fmt.Stringer, val interface{}) error
	RemoveItem(key fmt.Stringer) error
}
