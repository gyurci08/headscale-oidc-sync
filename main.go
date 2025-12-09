package main

import (
	"fmt"
	"os"
	"strings"

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

	users, err := ldapClient.QueryUsersWithGroups()
	if err != nil {
		log.Error("Failed to query LDAP users with roles", "error", err)
		os.Exit(1)
	}

	var filteredUsers []ldap.User
	for _, user := range users {
		var filteredUserGroups []ldap.Group
		isInGroup := false
		for _, group := range user.Groups {
			if strings.HasPrefix(group.Name, cfg.App.GroupPrefix) {
				isInGroup = true
				filteredUserGroups = append(filteredUserGroups, group)
			}
		}
		user.Groups = filteredUserGroups
		if isInGroup {
			filteredUsers = append(filteredUsers, user)
		}
	}

	for _, user := range filteredUsers {
		var identifier = user.Email
		if !strings.Contains(user.Email, "@") {
			identifier = user.Username + "@"
		}
		var groups []string
		for _, group := range user.Groups {
			groups = append(groups, group.Name)
		}
		log.Info("user", "email", identifier, "groups", groups)
	}
}
