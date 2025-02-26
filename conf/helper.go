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

func GetLifeAIConfiguration() *LifeAIConfiguration {
	return &configuration.LifeAIConfig
}

func GetAppName() string {
	return configuration.AppName
}

func GetJwtConfig() *JwtConfiguration {
	return &configuration.JwtConfig
}

func IsDebugMode() bool {
	return configuration.Env == "DEV"
}
