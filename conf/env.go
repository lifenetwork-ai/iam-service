package conf

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type RedisConfiguration struct {
	RedisAddress string `mapstructure:"REDIS_ADDRESS"`
	RedisTtl     string `mapstructure:"REDIS_TTL"`
}

type DatabaseConfiguration struct {
	DbUser     string `mapstructure:"DB_USER"`
	DbPassword string `mapstructure:"DB_PASSWORD"`
	DbHost     string `mapstructure:"DB_HOST"`
	DbPort     string `mapstructure:"DB_PORT"`
	DbName     string `mapstructure:"DB_NAME"`
}

type SecretConfiguration struct {
	Mnemonic   string `mapstructure:"MNEMONIC"`
	Passphrase string `mapstructure:"PASSPHRASE"`
	Salt       string `mapstructure:"SALT"`
}

type AdminAccountConfiguration struct {
	AdminEmail    string `mapstructure:"ADMIN_EMAIL"`
	AdminPassword string `mapstructure:"ADMIN_PASSWORD"`
}

type SecureGenomAPIConfiguration struct {
	SecureGenomAPIBaseURL string `mapstructure:"SECURE_GENOM_API_BASE_URL"`
}

type Configuration struct {
	Database          DatabaseConfiguration       `mapstructure:",squash"`
	Redis             RedisConfiguration          `mapstructure:",squash"`
	Secret            SecretConfiguration         `mapstructure:",squash"`
	AdminAccount      AdminAccountConfiguration   `mapstructure:",squash"`
	SecureGenomClient SecureGenomAPIConfiguration `mapstructure:",squash"`
	AppName           string                      `mapstructure:"APP_NAME"`
	AppPort           uint32                      `mapstructure:"APP_PORT"`
	Env               string                      `mapstructure:"ENV"`
	LogLevel          string                      `mapstructure:"LOG_LEVEL"`
	JWTSecret         string                      `mapstructure:"JWT_SECRET"`
}

var configuration Configuration

var defaultConfigurations = map[string]any{
	"REDIS_ADDRESS":             "localhost:6379",
	"REDIS_TTL":                 "60",
	"APP_PORT":                  "9090",
	"ENV_FILE":                  ".env",
	"ENV":                       "DEV",
	"LOG_LEVEL":                 "debug",
	"DB_USER":                   "db_master",
	"DB_PASSWORD":               "123456aA",
	"DB_HOST":                   "localhost",
	"DB_PORT":                   "5432",
	"DB_NAME":                   "human-network-iam",
	"MNEMONIC":                  "",
	"PASSPHRASE":                "",
	"SALT":                      "",
	"ADMIN_EMAIL":               "",
	"ADMIN_PASSWORD":            "",
	"JWT_SECRET":                "",
	"SECURE_GENOM_API_BASE_URL": "https://secure-genom.humannetwork.life",
}

// loadDefaultConfigs sets default values for critical configurations
func loadDefaultConfigs() {
	for configKey, configValue := range defaultConfigurations {
		viper.SetDefault(configKey, configValue)
	}
}

func init() {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env"
	}

	viper.SetConfigFile("./.env")
	viper.AutomaticEnv()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults for critical configurations
	loadDefaultConfigs()

	if err := viper.ReadInConfig(); err != nil {
		viper.SetConfigFile(fmt.Sprintf("../%s", envFile))
		if err := viper.ReadInConfig(); err != nil {
			log.Logger.Printf("Error reading config file \"%s\", %v", envFile, err)
		}
	}

	if err := viper.Unmarshal(&configuration); err != nil {
		log.Fatal().Err(err).Msgf("Error unmarshalling configuration %v", err)
	}

	log.Info().Msg("Configuration loaded successfully")
}

func GetConfiguration() *Configuration {
	return &configuration
}

func GetRedisConnectionURL() string {
	return configuration.Redis.RedisAddress
}
