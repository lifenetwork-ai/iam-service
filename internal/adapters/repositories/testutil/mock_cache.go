package testutil

import (
	"context"

	"github.com/lifenetwork-ai/iam-service/infrastructures/caching/types"
	"github.com/lifenetwork-ai/iam-service/internal/wire/instances"
)

// NewMockCache creates a new instance of MockCache
func NewMockCache() types.CacheRepository {
	return instances.CacheRepositoryInstance(context.Background())
}
