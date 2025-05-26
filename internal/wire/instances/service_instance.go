package instances

import (
	"sync"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

var (
	initServicesOnce sync.Once

	lifeAIInstance services.LifeAIService
	emailInstance  services.EmailService
	smsInstance    services.SMSService
	jwtInstance    services.JWTService
)

func initServices() {
	config := conf.GetConfiguration()

	lifeAIEndpoint := config.LifeAIConfig.BackendURL
	logger.GetLogger().Infof("Initializing service clients with LifeAI endpoint: %s", lifeAIEndpoint)

	lifeAIInstance = services.NewLifeAIService(lifeAIEndpoint)
	emailInstance = services.NewEmailService()
	smsInstance = services.NewSMSService()
	jwtInstance = services.NewJWTService(
		config.JwtConfig.Secret,
		config.JwtConfig.AccessLifetime,
		config.JwtConfig.RefreshLifetime,
	)
}

func LifeAIServiceInstance() services.LifeAIService {
	initServicesOnce.Do(initServices)
	return lifeAIInstance
}

func EmailServiceInstance() services.EmailService {
	initServicesOnce.Do(initServices)
	return emailInstance
}

func SMSServiceInstance() services.SMSService {
	initServicesOnce.Do(initServices)
	return smsInstance
}

func JWTServiceInstance() services.JWTService {
	initServicesOnce.Do(initServices)
	return jwtInstance
}
