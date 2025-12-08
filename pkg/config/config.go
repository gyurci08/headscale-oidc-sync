package config

import (
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	App  AppConfig  `validate:"required"`
	Log  LogConfig  `validate:"required"`
	Db   DbConfig   `validate:"required"`
	Smtp SmtpConfig `validate:"required"`
	Oidc OidcConfig `validate:"required"`
}

type AppConfig struct {
	Port        int    `validate:"omitempty,gt=0"`
	Env         string `validate:"omitempty,oneof=development test production"`
	FrontendUrl string `validate:"required,url"`
}

type LogConfig struct {
	Level  string `validate:"omitempty,oneof=debug info warn"`
	Format string `validate:"omitempty,oneof=json text console"`
}

type DbConfig struct {
	Host     string `validate:"required"`
	Port     int    `validate:"required,gt=0"`
	Name     string `validate:"required"`
	User     string `validate:"required"`
	Password string `validate:"required"`
	Sslmode  string `validate:"required"`
}

type SmtpConfig struct {
	Enabled     bool   `validate:"omitempty,oneof=true false"`
	Host        string `validate:"required"`
	Port        int    `validate:"required,gt=0"`
	User        string `validate:"omitempty"`
	Password    string `validate:"omitempty"`
	FromAddress string `validate:"required,email"`
	FromName    string `validate:"required"`
}

type OidcConfig struct {
	Issuer       string `validate:"required,url"`
	ClientId     string `validate:"required"`
	ClientSecret string `validate:"required"`
	RedirectUrl  string `validate:"required,url"`
	Scope        string `validate:"required"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Port:        getEnvInt("APP_PORT", 8080),
			Env:         getEnv("APP_ENV", "development"),
			FrontendUrl: getEnv("APP_FRONTEND_URL", "http://localhost:4200"),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "debug"),
			Format: getEnv("LOG_FORMAT", "console"),
		},
		Db: DbConfig{
			Host:     getEnv("DB_HOST", "db"),
			Port:     getEnvInt("DB_PORT", 5432),
			Name:     getEnv("DB_NAME", "app"),
			User:     getEnv("DB_USER", "app"),
			Password: getEnv("DB_PASSWORD", "12345678"),
			Sslmode:  getEnv("DB_SSLMODE", "disable"),
		},
		Smtp: SmtpConfig{
			Enabled:     getEnvBool("SMTP_ENABLED", false),
			Host:        getEnv("SMTP_HOST", "localhost"),
			Port:        getEnvInt("SMTP_PORT", 25),
			User:        getEnv("SMTP_USER", ""),
			Password:    getEnv("SMTP_PASSWORD", ""),
			FromAddress: getEnv("SMTP_FROM_ADDRESS", "noreply@localhost.local"),
			FromName:    getEnv("SMTP_FROM_NAME", "App"),
		},
		Oidc: OidcConfig{
			Issuer:       getEnv("OIDC_ISSUER", "https://oidc.example.com/application/o/app/"),
			ClientId:     getEnv("OIDC_CLIENT_ID", "id"),
			ClientSecret: getEnv("OIDC_CLIENT_SECRET", "secret"),
			RedirectUrl:  getEnv("OIDC_REDIRECT_URL", "https://app.example.com/oidc/callback"),
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
