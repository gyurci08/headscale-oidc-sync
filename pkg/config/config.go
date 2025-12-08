package config

import (
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	App  AppConfig  `validate:"required"`
	Log  LogConfig  `validate:"required"`
	Oidc OidcConfig `validate:"required"`
}

type AppConfig struct {
	Port int    `validate:"omitempty,gt=0"`
	Env  string `validate:"omitempty,oneof=development test production"`
}

type LogConfig struct {
	Level  string `validate:"omitempty,oneof=debug info warn"`
	Format string `validate:"omitempty,oneof=json text console"`
}

type OidcConfig struct {
	Issuer       string `validate:"required,url"`
	ClientId     string `validate:"required"`
	ClientSecret string `validate:"required"`
	Scope        string `validate:"required"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Port: getEnvInt("APP_PORT", 8080),
			Env:  getEnv("APP_ENV", "development"),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "debug"),
			Format: getEnv("LOG_FORMAT", "console"),
		},
		Oidc: OidcConfig{
			Issuer:       getEnv("OIDC_ISSUER", "https://oidc.example.com/application/o/app/"),
			ClientId:     getEnv("OIDC_CLIENT_ID", "id"),
			ClientSecret: getEnv("OIDC_CLIENT_SECRET", "secret"),
			Scope:        getEnv("OIDC_SCOPE", "openid email profile"),
		},
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return fallback
}
