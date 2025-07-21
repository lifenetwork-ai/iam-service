package instances

import (
	"context"
	"sync"

	"github.com/lifenetwork-ai/iam-service/conf"
	queue "github.com/lifenetwork-ai/iam-service/infrastructures/otp_queue"
	queuetypes "github.com/lifenetwork-ai/iam-service/infrastructures/otp_queue/types"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
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
			otpQueueRepo = queue.NewRedisOTPQueueRepository(RedisClientInstance())
		default:
			logger.GetLogger().Info("Using in-memory cache for OTP queue")
			otpQueueRepo = queue.NewMemoryOTPQueueRepository(GoCacheClientInstance())
		}
	})
	return otpQueueRepo
}
