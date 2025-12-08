package main

import (
	"fmt"
	"os"

	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/config"
	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/ldap"
	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/logger"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.NewLogger(*cfg, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}

	log.Debug("Starting logs...")

	ldapClient, err := ldap.NewClient(cfg.Ldap, log)
	if err != nil {
		log.Error("Failed to create LDAP client", "error", err)
		os.Exit(1)
	}
	defer ldapClient.Close()

	users, err := ldapClient.QueryUsers()
	if err != nil {
		log.Error("Failed to query LDAP users", "error", err)
		os.Exit(1)
	}

	log.Info("LDAP users:", "users", users)

	groups, err := ldapClient.QueryGroups()
	if err != nil {
		log.Error("Failed to query LDAP groups", "error", err)
		os.Exit(1)
	}

	log.Info("LDAP groups:", "groups", groups)
}
