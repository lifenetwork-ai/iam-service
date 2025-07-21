package instances

import (
	"context"
	"sync"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/constants"
	queue "github.com/lifenetwork-ai/iam-service/infrastructures/otp_queue"
	queuetypes "github.com/lifenetwork-ai/iam-service/infrastructures/otp_queue/types"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
)

var (
	otpQueueOnce sync.Once
	otpQueueRepo queuetypes.OTPQueueRepository
)

// OTPQueueRepositoryInstance returns a singleton instance of OTPQueueRepository
func OTPQueueRepositoryInstance(ctx context.Context) queuetypes.OTPQueueRepository {
	otpQueueOnce.Do(func() {
		cacheType := conf.GetCacheType()
		switch cacheType {
		case "redis":
			logger.GetLogger().Info("Using Redis for OTP queue")
			config := conf.GetRedisConfiguration()
			redisClient := redis.NewClient(&redis.Options{
				Addr: config.RedisAddress,
			})
			otpQueueRepo = queue.NewRedisOTPQueueRepository(redisClient)
		default:
			logger.GetLogger().Info("Using in-memory cache for OTP queue")
			memCache := cache.New(constants.DefaultExpiration, constants.CleanupInterval)
			otpQueueRepo = queue.NewMemoryOTPQueueRepository(memCache)
		}
	})
	return otpQueueRepo
}
