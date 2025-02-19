package providers

import (
	"sync"

	"github.com/genefriendway/human-network-iam/conf"
	"github.com/genefriendway/human-network-iam/infrastructures/clients"
	"github.com/genefriendway/human-network-iam/packages/logger"
)

var (
	lifeAIOnce     sync.Once
	lifeAIInstance *clients.LifeAIClient
)

// ProvideLifeAIClient provides a singleton instance of LifeAIClient.
func ProvideLifeAIClient() *clients.LifeAIClient {
	lifeAIOnce.Do(func() {
		// Get LifeAI endpoint from config
		config := conf.GetConfiguration()
		lifeAIEndpoint := config.LifeAIConfig.BackendURL

		logger.GetLogger().Infof("Initializing LifeAI client with endpoint: %s", lifeAIEndpoint)

		// Create LifeAI client
		lifeAIInstance = clients.NewLifeAIClient(lifeAIEndpoint)
	})

	return lifeAIInstance
}
