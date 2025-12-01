package conf

type KetoConfiguration struct {
	DefaultReadURL  string `mapstructure:"KETO_DEFAULT_READ_URL"`
	DefaultWriteURL string `mapstructure:"KETO_DEFAULT_WRITE_URL"`
}

func GetKetoConfig() *KetoConfiguration {
	return &configuration.Keto
}
