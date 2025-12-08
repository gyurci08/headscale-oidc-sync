package config

type LogConfig struct {
	Level  string `validate:"omitempty,oneof=debug info warn"`
	Format string `validate:"omitempty,oneof=json text console"`
}

func NewLogConfig() LogConfig {
	return LogConfig{
		Level:  getEnv("LOG_LEVEL", "debug"),
		Format: getEnv("LOG_FORMAT", "console"),
	}
}
