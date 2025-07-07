package conf

import "github.com/spf13/viper"

type KratosConfig struct {
	PublicEndpoint string
	AdminEndpoint  string
}

func LoadKratosConfig() *KratosConfig {
	return &KratosConfig{
		PublicEndpoint: viper.GetString("kratos.public_endpoint"),
		AdminEndpoint:  viper.GetString("kratos.admin_endpoint"),
	}
}
