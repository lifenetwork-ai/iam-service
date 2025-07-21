package instances

import (
	"sync"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/constants"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
)

var (
	goCacheOnce     sync.Once
	goCacheInstance *cache.Cache

	redisOnce       sync.Once
	redisClientInst *redis.Client
)

func GoCacheClientInstance() *cache.Cache {
	goCacheOnce.Do(func() {
		logger.GetLogger().Info("Initializing GoCache instance")
		goCacheInstance = cache.New(constants.DefaultExpiration, constants.CleanupInterval)
	})
	return goCacheInstance
}

func RedisClientInstance() *redis.Client {
	redisOnce.Do(func() {
		config := conf.GetRedisConfiguration()
		logger.GetLogger().Infof("Initializing Redis client at %s", config.RedisAddress)
		redisClientInst = redis.NewClient(&redis.Options{
			Addr: config.RedisAddress,
		})
	})
	return redisClientInst
}
