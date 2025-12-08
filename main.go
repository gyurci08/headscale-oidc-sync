package main

import (
	"context"

	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/config"
	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/logger"
	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/oidc"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	log, err := logger.NewLogger(*cfg, nil)
	if err != nil {
		panic(err)
	}

	oidc, err := oidc.NewOidcClient(*cfg, log)
	if err != nil {
		log.Error("Could not create oidc client!")
		panic(err)
	}

	groups, err := oidc.ListGroups(context.Background())
	if err != nil {
		log.Error("Could not list groups!")
		panic(err)
	}
	for _, group := range groups {
		log.Info(group)
	}

	log.Info("Application started...")
}
