package config

type AppConfig struct {
	Port        int    `validate:"omitempty,gt=0"`
	Env         string `validate:"omitempty,oneof=development test production"`
	GroupPrefix string `validate:"required"`
}

func NewAppConfig() AppConfig {
	return AppConfig{
		Port:        getEnvInt("APP_PORT", 8080),
		Env:         getEnvValue("APP_ENV", "production"),
		GroupPrefix: getEnvValue("APP_GROUP_PREFIX", ""),
	}
}
