package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/robfig/cron/v3"
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
	log.Info("Configuration loaded successfully")

	log.Info("Running initial sync...")
	syncACL(cfg, log)

	// Start cron scheduler
	c := cron.New()
	schedule := cfg.App.CronSchedule
	_, err = c.AddFunc(schedule, func() {
		log.Debug("Cron job triggered, running sync...")
		syncACL(cfg, log)
	})
	if err != nil {
		log.Error("Failed to add cron job", "error", err)
		os.Exit(1)
	}

	log.Info("Cron scheduler started", "schedule", schedule)
	c.Start()
	defer c.Stop()

	// Keep the main function running
	log.Info("Application is running and waiting for scheduled jobs...")
	select {}
}

func syncACL(cfg *config.Config, log logger.ILogger) {
	log.Debug("Starting LDAP client setup...")
	ldapClient, err := ldap.NewClient(cfg.Ldap, log)
	if err != nil {
		log.Error("Failed to create LDAP client", "error", err)
		return
	}
	defer ldapClient.Close()
	log.Debug("LDAP client setup complete")

	log.Info("Querying LDAP users with groups...")
	users, err := ldapClient.QueryUsersWithGroups()
	if err != nil {
		log.Error("Failed to query LDAP users with roles", "error", err)
		return
	}
	log.Info("LDAP query complete", "total_users", len(users))

	// Load existing ACL file
	aclFilePath := cfg.App.AclJson
	log.Debug("Reading existing ACL file", "path", aclFilePath)
	aclData, err := os.ReadFile(aclFilePath)
	if err != nil {
		log.Error("Failed to read ACL file", "path", aclFilePath, "error", err)
		return
	}

	// Parse the ACL file
	var existingACL ACL
	log.Debug("Parsing existing ACL file")
	if err := json.Unmarshal(aclData, &existingACL); err != nil {
		log.Error("Failed to parse existing ACL file", "path", aclFilePath, "error", err)
		return
	}

	// Generate new groups from LDAP
	log.Debug("Generating new groups from LDAP data")
	newGroups := generateGroupsFromLDAP(users, cfg.App.GroupPrefix)

	// Create updated structure preserving original acls as raw JSON
	updatedACL := ACL{
		Groups: newGroups,
		ACLs:   existingACL.ACLs,
	}

	// Marshal updated file
	log.Debug("Marshaling updated ACL file")
	updatedJSON, err := json.MarshalIndent(updatedACL, "", "  ")
	if err != nil {
		log.Error("Failed to marshal updated ACL file", "error", err)
		return
	}

	// Check if ACL content has changed
	if string(updatedJSON) != string(aclData) {
		log.Debug("ACL content changed, updating file...")
		if err := os.WriteFile(aclFilePath, updatedJSON, 0644); err != nil {
			log.Error("Failed to write ACL file", "path", aclFilePath, "error", err)
			return
		}

		log.Info("ACL file updated successfully",
			"path", aclFilePath,
			"total_groups", len(newGroups),
			"total_users_in_groups", countUniqueUsersInGroups(newGroups))

		// Reload headscale container if enabled in config
		if cfg.App.IsReloadHeadscale {
			containerName := cfg.App.HeadscaleContainerName
			log.Debug("Reloading headscale container", "container", containerName)
			cmd := exec.Command("docker", "kill", "--signal=HUP", containerName)
			if err := cmd.Run(); err != nil {
				log.Error("Failed to reload headscale container", "error", err)
			} else {
				log.Info("Headscale container reloaded", "container", containerName)
			}
		} else {
			log.Info("Headscale reload disabled in config")
		}
	} else {
		log.Info("ACL file unchanged, no reload needed")
	}
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
