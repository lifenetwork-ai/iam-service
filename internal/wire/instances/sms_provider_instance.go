package instances

import (
	"sync"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms"
	services "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/services"
)

var (
	smsProviderOnce sync.Once
	smsProvider     services.SMSProvider
)

func SMSProviderInstance() services.SMSProvider {
	smsProviderOnce.Do(func() {
		smsProvider = sms.NewSMSProvider(conf.GetSmsConfiguration())
	})
	return smsProvider
}
