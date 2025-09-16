package conf

func GetConfiguration() *Configuration {
	return &configuration
}

func GetRedisConfiguration() *RedisConfiguration {
	return &configuration.Redis
}

func GetCacheType() string {
	return configuration.CacheType
}

func GetAppName() string {
	return configuration.AppName
}

func GetSmsConfiguration() *SmsConfiguration {
	return &configuration.Sms
}

func IsDebugMode() bool {
	return configuration.Env == "DEV"
}

func IsDevReviewerBypassEnabled() bool {
	return configuration.DevReviewer.DevReviewerByPass
}

func DevReviewerMagicOTP() string {
	if v := configuration.DevReviewer.DevReviewerMagicOTP; v != "" {
		return v
	}
	return "123456"
}

func DevReviewerIdentifier() string {
	return configuration.DevReviewer.DevReviewerIdentifier
}

func GetMockWebhookURL() string {
	return configuration.MockWebhookURL
}
