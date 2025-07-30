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

type RootAccountConfiguration struct {
	RootUsername string `mapstructure:"IAM_ROOT_USERNAME"`
	RootPassword string `mapstructure:"IAM_ROOT_PASSWORD"`
}

type Configuration struct {
	Database     DatabaseConfiguration    `mapstructure:",squash"`
	Redis        RedisConfiguration       `mapstructure:",squash"`
	RootAccount  RootAccountConfiguration `mapstructure:",squash"`
	AppName      string                   `mapstructure:"APP_NAME"`
	AppPort      uint32                   `mapstructure:"APP_PORT"`
	Env          string                   `mapstructure:"ENV"`
	LogLevel     string                   `mapstructure:"LOG_LEVEL"`
	CacheType    string                   `mapstructure:"CACHE_TYPE"`
	KratosConfig KratosConfiguration      `mapstructure:",squash"`
	Keto         KetoConfiguration        `mapstructure:",squash"`
	Sms          SmsConfiguration         `mapstructure:",squash"`
}

type TwilioConfiguration struct {
	TwilioAccountSID string `mapstructure:"TWILIO_ACCOUNT_SID"`
	TwilioAuthToken  string `mapstructure:"TWILIO_AUTH_TOKEN"`
	TwilioFrom       string `mapstructure:"TWILIO_FROM"`
}

type WhatsappConfiguration struct {
	WhatsappPhoneID     string `mapstructure:"WHATSAPP_PHONE_ID"`
	WhatsappAccessToken string `mapstructure:"WHATSAPP_ACCESS_TOKEN"`
}

type SmsConfiguration struct {
	Twilio   TwilioConfiguration   `mapstructure:",squash"`
	Whatsapp WhatsappConfiguration `mapstructure:",squash"`
}

var configuration Configuration

// NOTE: when adding a new env, you need to add it to the defaultConfigurations map
// TODO: add a way remove this behavior
var defaultConfigurations = map[string]any{
	"REDIS_ADDRESS":                  "localhost:6379",
	"REDIS_TTL":                      "60",
	"APP_PORT":                       "8080",
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
	"IAM_ROOT_USERNAME":              "",
	"IAM_ROOT_PASSWORD":              "",
	"KETO_DEFAULT_READ_URL":          "",
	"KETO_DEFAULT_WRITE_URL":         "",
	"MOCK_WEBHOOK_URL":               "",
	"TWILIO_ACCOUNT_SID":             "",
	"TWILIO_AUTH_TOKEN":              "",
	"TWILIO_FROM":                    "",
	"WHATSAPP_PHONE_ID":              "",
	"WHATSAPP_ACCESS_TOKEN":          "",
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

	viper.AutomaticEnv()

	// Set default values after AutomaticEnv
	loadDefaultConfigs()

	// Attempt to read the .env file
	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Loaded configuration from file: %s", envFile)
	} else {
		log.Printf("Config file \"%s\" not found or unreadable, using environment variables and defaults", envFile)
	}

	// Unmarshal values into the global `configuration` struct
	if err := viper.Unmarshal(&configuration); err != nil {
		log.Fatalf("Error unmarshalling configuration: %v", err)
	}

	log.Println("Configuration loaded successfully")
}
