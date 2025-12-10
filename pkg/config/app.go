package config

type AppConfig struct {
	Port                   int    `validate:"omitempty,gt=0"`
	Env                    string `validate:"omitempty,oneof=development test production"`
	GroupPrefix            string `validate:"required"`
	AclJson                string `validate:"required"`
	IsReloadHeadscale      bool
	HeadscaleContainerName string
	CronSchedule           string `validate:"omitempty,cron"`
}

func NewAppConfig() AppConfig {
	return AppConfig{
		Port:                   getEnvInt("APP_PORT", 8080),
		Env:                    getEnvValue("APP_ENV", "production"),
		GroupPrefix:            getEnvValue("APP_GROUP_PREFIX", ""),
		AclJson:                getEnvValue("APP_ACL_JSON", ""),
		IsReloadHeadscale:      getEnvBool("APP_IS_RELOAD_HEADSCALE", false),
		HeadscaleContainerName: getEnvValue("APP_HEADSCALE_CONTAINER_NAME", "headscale"),
		CronSchedule:           getEnvValue("APP_CRON_SCHEDULE", "@every 1h"),
	}
}
