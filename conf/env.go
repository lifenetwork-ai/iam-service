package conf

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type RedisConfiguration struct {
	RedisAddress string `mapstructure:"REDIS_ADDRESS"`
	RedisTtl     string `mapstructure:"REDIS_TTL"`
}

type DatabaseConfiguration struct {
	DbUser                    string `mapstructure:"DB_USER"`
	DbPassword                string `mapstructure:"DB_PASSWORD"`
	DbHost                    string `mapstructure:"DB_HOST"`
	DbPort                    string `mapstructure:"DB_PORT"`
	DbName                    string `mapstructure:"DB_NAME"`
	DbMaxOpenConns            string `mapstructure:"DB_MAX_OPEN_CONNS"`
	DbMaxIdleConns            string `mapstructure:"DB_MAX_IDLE_CONNS"`
	DbConnMaxLifetimeInMinute string `mapstructure:"DB_CONN_MAX_LIFETIME_IN_MINUTE"`
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

type LifeAIConfiguration struct {
	BackendURL string `mapstructure:"LIFE_AI_BACKEND_URL"`
}

type EmailConfiguration struct {
	EmailHost     string `mapstructure:"EMAIL_HOST"`
	EmailPort     string `mapstructure:"EMAIL_PORT"`
	EmailUsername string `mapstructure:"EMAIL_USERNAME"`
	EmailPassword string `mapstructure:"EMAIL_PASSWORD"`
}

type SmsConfiguration struct {
	SmsProvider string `mapstructure:"SMS_PROVIDER"`
	SmsUsername string `mapstructure:"SMS_USERNAME"`
	SmsPassword string `mapstructure:"SMS_PASSWORD"`
}

type JwtConfiguration struct {
	Secret          string `mapstructure:"JWT_SECRET"`
	AccessLifetime  int64  `mapstructure:"JWT_ACCESS_TOKEN_LIFETIME"`  // second
	RefreshLifetime int64  `mapstructure:"JWT_REFRESH_TOKEN_LIFETIME"` // second
}

type Configuration struct {
	Database     DatabaseConfiguration     `mapstructure:",squash"`
	Redis        RedisConfiguration        `mapstructure:",squash"`
	Secret       SecretConfiguration       `mapstructure:",squash"`
	AdminAccount AdminAccountConfiguration `mapstructure:",squash"`
	AppName      string                    `mapstructure:"APP_NAME"`
	AppPort      uint32                    `mapstructure:"APP_PORT"`
	Env          string                    `mapstructure:"ENV"`
	LogLevel     string                    `mapstructure:"LOG_LEVEL"`
	JWTSecret    string                    `mapstructure:"JWT_SECRET"`
	LifeAIConfig LifeAIConfiguration       `mapstructure:",squash"`
	CacheType    string                    `mapstructure:"CACHE_TYPE"`
	EmailConfig  EmailConfiguration        `mapstructure:",squash"`
	SmsConfig    SmsConfiguration          `mapstructure:",squash"`
	JwtConfig    JwtConfiguration          `mapstructure:",squash"`
}

var configuration Configuration

var defaultConfigurations = map[string]any{
	"REDIS_ADDRESS":                  "localhost:6379",
	"REDIS_TTL":                      "60",
	"APP_PORT":                       "9090",
	"ENV_FILE":                       ".env",
	"ENV":                            "DEV",
	"LOG_LEVEL":                      "debug",
	"DB_USER":                        "db_master",
	"DB_PASSWORD":                    "123456aA",
	"DB_HOST":                        "localhost",
	"DB_PORT":                        "5432",
	"DB_NAME":                        "human-network-iam",
	"DB_MAX_IDLE_CONNS":              "5",
	"DB_MAX_OPEN_CONNS":              "15",
	"DB_CONN_MAX_LIFETIME_IN_MINUTE": "60",
	"MNEMONIC":                       "",
	"PASSPHRASE":                     "",
	"SALT":                           "",
	"JWT_SECRET":                     "Abc@13579",
	"JWT_ACCESS_TOKEN_LIFETIME":      "3600",  // 1 hour
	"JWT_REFRESH_TOKEN_LIFETIME":     "86400", // 24 hours
	"LIFE_AI_BACKEND_URL":            "https://nightly.lifenetwork.ai",
}

// loadDefaultConfigs sets default values for critical configurations
func loadDefaultConfigs() {
	for configKey, configValue := range defaultConfigurations {
		viper.SetDefault(configKey, configValue)
	}
}

func init() {
	// Set environment variable for .env file location
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env" // Default to .env if ENV_FILE is not set
	}

	// Set Viper to look for the config file
	viper.SetConfigFile(envFile)
	viper.SetConfigType("env")                             // Explicitly tell Viper it's an .env file
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // Replace dots with underscores

	// Set default values before reading config
	loadDefaultConfigs()

	// Attempt to read the .env file
	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Loaded configuration from file: %s", envFile)
	} else {
		viper.AutomaticEnv() // Enable reading from environment variables
		log.Printf("Config file \"%s\" not found or unreadable, falling back to environment variables", envFile)
	}

	// Unmarshal values into the global `configuration` struct
	if err := viper.Unmarshal(&configuration); err != nil {
		log.Fatalf("Error unmarshalling configuration: %v", err)
	}

	log.Println("Configuration loaded successfully")
}
