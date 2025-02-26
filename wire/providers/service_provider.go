package providers

import (
	"sync"

	"github.com/genefriendway/human-network-iam/conf"
	"github.com/genefriendway/human-network-iam/internal/adapters/services"
	"github.com/genefriendway/human-network-iam/packages/logger"
)

var (
	serviceOnce    sync.Once
	lifeAIInstance services.LifeAIService
	emailInstance  services.EmailService
	smsInstance    services.SMSService
	jwtInstance    services.JWTService
)

// ProvideLifeAIService provides a singleton instance of LifeAIClient.
func ProvideLifeAIService() services.LifeAIService {
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

// ProvideEmailService provides a singleton instance of EmailService.
func ProvideEmailService() services.EmailService {
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

// ProvideSMSService provides a singleton instance of SMSService.
func ProvideSMSService() services.SMSService {
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

// ProvideJWTService provides a singleton instance of JWTService.
func ProvideJWTService() services.JWTService {
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
