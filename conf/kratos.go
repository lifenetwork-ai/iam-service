package conf

type KratosConfiguration struct{}

func GetKratosConfig() *KratosConfiguration {
	return &configuration.KratosConfig
}
