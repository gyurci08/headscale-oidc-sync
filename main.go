package main

import (
	"os"

	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/config"
	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/logger"
)

func setupLogger(cfg *config.Config) (logger.ILogger, error) {
	loggerParams := logger.NewParams{
		Config: *cfg,
		Writer: os.Stdout,
	}
	return logger.NewLogger(loggerParams)
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	log, err := setupLogger(cfg)
	if err != nil {
		panic(err)
	}

	log.Info("Application started...")
}
