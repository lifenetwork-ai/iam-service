package testutil

import (
	"context"

	"github.com/lifenetwork-ai/iam-service/infrastructures/interfaces"
	"github.com/lifenetwork-ai/iam-service/internal/wire/instances"
)

// NewMockCache creates a new instance of MockCache
func NewMockCache() interfaces.CacheRepository {
	return instances.CacheRepositoryInstance(context.Background())
}
