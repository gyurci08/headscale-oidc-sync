package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/config"
	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/ldap"
	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/logger"
)

// Define a struct that matches the ACL file structure
type ACL struct {
	Groups map[string][]string `json:"groups"`
	ACLs   json.RawMessage     `json:"acls"`
}

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

	log.Info("LDAP query complete", "total_users", len(users))

	// Load existing ACL file
	aclFilePath := cfg.App.AclJson
	aclData, err := os.ReadFile(aclFilePath)
	if err != nil {
		log.Error("Failed to read ACL file", "path", aclFilePath, "error", err)
		os.Exit(1)
	}

	// Parse the ACL file
	var existingACL ACL
	if err := json.Unmarshal(aclData, &existingACL); err != nil {
		log.Error("Failed to parse existing ACL file", "path", aclFilePath, "error", err)
		os.Exit(1)
	}

	// Generate new groups from LDAP
	newGroups := generateGroupsFromLDAP(users, cfg.App.GroupPrefix)

	// Create updated structure preserving original acls as raw JSON
	updatedACL := ACL{
		Groups: newGroups,
		ACLs:   existingACL.ACLs,
	}

	// Marshal updated file
	updatedJSON, err := json.MarshalIndent(updatedACL, "", "  ")
	if err != nil {
		log.Error("Failed to marshal updated ACL file", "error", err)
		os.Exit(1)
	}

	// Write back to file
	if err := os.WriteFile(aclFilePath, updatedJSON, 0644); err != nil {
		log.Error("Failed to write ACL file", "path", aclFilePath, "error", err)
		os.Exit(1)
	}

	log.Info("ACL file updated successfully",
		"path", aclFilePath,
		"total_groups", len(newGroups),
		"total_users_in_groups", countUniqueUsersInGroups(newGroups))
}

// generateGroupsFromLDAP creates the groups map from LDAP data
func generateGroupsFromLDAP(users []ldap.User, groupPrefix string) map[string][]string {
	groupMap := make(map[string][]string)

	for _, user := range users {
		for _, group := range user.Groups {
			if strings.HasPrefix(group.Name, groupPrefix) {
				identifier := user.Email
				if !strings.Contains(user.Email, "@") {
					identifier = user.Username + "@"
				}
				key := "group:" + group.Name
				groupMap[key] = append(groupMap[key], identifier)
			}
		}
	}

	return groupMap
}

// countUniqueUsersInGroups returns the total number of unique users across all groups
func countUniqueUsersInGroups(groups map[string][]string) int {
	userSet := make(map[string]bool)

	for _, userEmails := range groups {
		for _, email := range userEmails {
			userSet[email] = true
		}
	}

	return len(userSet)
}
