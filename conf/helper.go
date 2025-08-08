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

func GetDbEncryptionKey() [32]byte {
	key := [32]byte{}
	copy(key[:], []byte(configuration.DbEncryptionKey))
	return key
}
