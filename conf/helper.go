package conf

func GetConfiguration() *Configuration {
	return &configuration
}

func GetKratosConfig() *KratosConfiguration {
	return &configuration.KratosConfig
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

func IsDebugMode() bool {
	return configuration.Env == "DEV"
}
