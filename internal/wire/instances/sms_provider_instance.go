package instances

import (
	"sync"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

var (
	smsServiceOnce sync.Once
	smsService     *sms.SMSService
	smsServiceErr  error
)

// SMSProviderInstance returns a singleton instance of the SMS service
func SMSServiceInstance(zaloTokenRepo domainrepo.ZaloTokenRepository) *sms.SMSService {
	smsServiceOnce.Do(func() {
		service, err := sms.NewSMSService(conf.GetSmsConfiguration(), zaloTokenRepo)
		if err != nil {
			logger.GetLogger().Errorf("Failed to create SMS service: %v", err)
			smsServiceErr = err
			return
		}
		smsService = service
	})

	if smsServiceErr != nil {
		logger.GetLogger().Errorf("SMS service initialization failed: %v", smsServiceErr)
		return nil
	}

	return smsService
}
