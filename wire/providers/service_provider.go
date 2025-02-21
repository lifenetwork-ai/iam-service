package providers

import (
	"sync"

	"github.com/genefriendway/human-network-iam/conf"
	email_service "github.com/genefriendway/human-network-iam/internal/adapters/services/email"
	lifeai_service "github.com/genefriendway/human-network-iam/internal/adapters/services/lifeai"
	sms_service "github.com/genefriendway/human-network-iam/internal/adapters/services/sms"
	"github.com/genefriendway/human-network-iam/packages/logger"
)

var (
	serviceOnce    sync.Once
	lifeAIInstance lifeai_service.LifeAIService
	emailInstance  email_service.EmailService
	smsInstance    sms_service.SMSService
)

// ProvideLifeAIService provides a singleton instance of LifeAIClient.
func ProvideLifeAIService() lifeai_service.LifeAIService {
	serviceOnce.Do(func() {
		// Get LifeAI endpoint from config
		config := conf.GetConfiguration()
		lifeAIEndpoint := config.LifeAIConfig.BackendURL

		logger.GetLogger().Infof("Initializing LifeAI client with endpoint: %s", lifeAIEndpoint)

		// Create service instances
		lifeAIInstance = lifeai_service.NewLifeAIService(lifeAIEndpoint)
		emailInstance = email_service.NewEmailService()
		smsInstance = sms_service.NewSMSService()
	})

	return lifeAIInstance
}

// ProvideEmailService provides a singleton instance of EmailService.
func ProvideEmailService() email_service.EmailService {
	serviceOnce.Do(func() {
		// Get email service configuration from config
		config := conf.GetConfiguration()
		lifeAIEndpoint := config.LifeAIConfig.BackendURL

		logger.GetLogger().Infof("Initializing LifeAI client with endpoint: %s", lifeAIEndpoint)

		// Create service instances
		lifeAIInstance = lifeai_service.NewLifeAIService(lifeAIEndpoint)
		emailInstance = email_service.NewEmailService()
		smsInstance = sms_service.NewSMSService()
	})

	return emailInstance
}

// ProvideSMSService provides a singleton instance of SMSService.
func ProvideSMSService() sms_service.SMSService {
	serviceOnce.Do(func() {
		// Get SMS service configuration from config
		config := conf.GetConfiguration()
		lifeAIEndpoint := config.LifeAIConfig.BackendURL

		logger.GetLogger().Infof("Initializing LifeAI client with endpoint: %s", lifeAIEndpoint)

		// Create service instances
		lifeAIInstance = lifeai_service.NewLifeAIService(lifeAIEndpoint)
		emailInstance = email_service.NewEmailService()
		smsInstance = sms_service.NewSMSService()
	})

	return smsInstance
}
