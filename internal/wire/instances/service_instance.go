package instances

import (
	"sync"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

var (
	serviceOnce    sync.Once
	lifeAIInstance services.LifeAIService
	emailInstance  services.EmailService
	smsInstance    services.SMSService
	jwtInstance    services.JWTService
)

// LifeAIServiceInstance provides a singleton instance of LifeAIClient.
func LifeAIServiceInstance() services.LifeAIService {
	serviceOnce.Do(func() {
		// Get LifeAI endpoint from config
		config := conf.GetConfiguration()
		lifeAIEndpoint := config.LifeAIConfig.BackendURL

		logger.GetLogger().Infof("Initializing LifeAI client with endpoint: %s", lifeAIEndpoint)

		// Create service instances
		lifeAIInstance = services.NewLifeAIService(lifeAIEndpoint)
		emailInstance = services.NewEmailService()
		smsInstance = services.NewSMSService()
		jwtInstance = services.NewJWTService(
			config.JwtConfig.Secret,
			config.JwtConfig.AccessLifetime,
			config.JwtConfig.RefreshLifetime,
		)
	})

	return lifeAIInstance
}

// EmailServiceInstance provides a singleton instance of EmailService.
func EmailServiceInstance() services.EmailService {
	serviceOnce.Do(func() {
		// Get email service configuration from config
		config := conf.GetConfiguration()
		lifeAIEndpoint := config.LifeAIConfig.BackendURL

		logger.GetLogger().Infof("Initializing LifeAI client with endpoint: %s", lifeAIEndpoint)

		// Create service instances
		lifeAIInstance = services.NewLifeAIService(lifeAIEndpoint)
		emailInstance = services.NewEmailService()
		smsInstance = services.NewSMSService()
		jwtInstance = services.NewJWTService(
			config.JwtConfig.Secret,
			config.JwtConfig.AccessLifetime,
			config.JwtConfig.RefreshLifetime,
		)
	})

	return emailInstance
}

// SMSServiceInstance provides a singleton instance of SMSService.
func SMSServiceInstance() services.SMSService {
	serviceOnce.Do(func() {
		// Get SMS service configuration from config
		config := conf.GetConfiguration()
		lifeAIEndpoint := config.LifeAIConfig.BackendURL

		logger.GetLogger().Infof("Initializing LifeAI client with endpoint: %s", lifeAIEndpoint)

		// Create service instances
		lifeAIInstance = services.NewLifeAIService(lifeAIEndpoint)
		emailInstance = services.NewEmailService()
		smsInstance = services.NewSMSService()
		jwtInstance = services.NewJWTService(
			config.JwtConfig.Secret,
			config.JwtConfig.AccessLifetime,
			config.JwtConfig.RefreshLifetime,
		)
	})

	return smsInstance
}

// JWTServiceInstance provides a singleton instance of JWTService.
func JWTServiceInstance() services.JWTService {
	serviceOnce.Do(func() {
		// Get JWT service configuration from config
		config := conf.GetConfiguration()
		lifeAIEndpoint := config.LifeAIConfig.BackendURL

		logger.GetLogger().Infof("Initializing LifeAI client with endpoint: %s", lifeAIEndpoint)

		// Create service instances
		lifeAIInstance = services.NewLifeAIService(lifeAIEndpoint)
		emailInstance = services.NewEmailService()
		smsInstance = services.NewSMSService()
		jwtInstance = services.NewJWTService(
			config.JwtConfig.Secret,
			config.JwtConfig.AccessLifetime,
			config.JwtConfig.RefreshLifetime,
		)
	})

	return jwtInstance
}
