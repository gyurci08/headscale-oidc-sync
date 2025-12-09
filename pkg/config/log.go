package config

type LogConfig struct {
	Level  string `validate:"omitempty,oneof=debug info warn"`
	Format string `validate:"omitempty,oneof=json text console"`
}

func NewLogConfig() LogConfig {
	return LogConfig{
		Level:  getEnvValue("LOG_LEVEL", "info"),
		Format: getEnvValue("LOG_FORMAT", "console"),
	}
}
