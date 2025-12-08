package ldap

import (
	"fmt"

	"github.com/go-ldap/ldap/v3"
	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/config"
	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/logger"
)

type Client struct {
	conn   *ldap.Conn
	config config.LdapConfig
	log    logger.ILogger
}

func NewClient(cfg config.LdapConfig, log logger.ILogger) (*Client, error) {
	var url string
	switch cfg.Protocol {
	case "ssl", "tls":
		url = fmt.Sprintf("ldaps://%s:%d", cfg.Host, cfg.Port)
	case "starttls":
		url = fmt.Sprintf("ldap://%s:%d", cfg.Host, cfg.Port)
	default:
		url = fmt.Sprintf("ldap://%s:%d", cfg.Host, cfg.Port)
	}

	conn, err := ldap.DialURL(url)
	if err != nil {
		log.Error("Failed to connect to LDAP", "error", err)
		return nil, err
	}

	if cfg.Protocol == "starttls" {
		err = conn.StartTLS(nil)
		if err != nil {
			log.Error("Failed to start TLS", "error", err)
			return nil, err
		}
	}

	err = conn.Bind(cfg.BindDN, cfg.BindPW)
	if err != nil {
		log.Error("Failed to bind to LDAP", "error", err)
		return nil, err
	}

	log.Debug("Connected to LDAP")

	return &Client{conn: conn, config: cfg, log: log}, nil
}

func (c *Client) Close() {
	c.conn.Close()
	c.log.Debug("LDAP connection closed")
}

func (c *Client) QueryUsers() ([]string, error) {
	searchRequest := ldap.NewSearchRequest(
		c.config.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		c.config.UserFilter,
		[]string{"mail"},
		nil,
	)

	sr, err := c.conn.Search(searchRequest)
	if err != nil {
		c.log.Error("Failed to query users", "error", err)
		return nil, err
	}

	var emails []string
	for _, entry := range sr.Entries {
		email := entry.GetAttributeValue("mail")
		emails = append(emails, email)
	}

	c.log.Info("Queried users", "count", len(emails))
	return emails, nil
}

func (c *Client) QueryGroups() ([]string, error) {
	searchRequest := ldap.NewSearchRequest(
		c.config.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		c.config.GroupFilter,
		[]string{"cn"},
		nil,
	)

	sr, err := c.conn.Search(searchRequest)
	if err != nil {
		c.log.Error("Failed to query groups", "error", err)
		return nil, err
	}

	var groups []string
	for _, entry := range sr.Entries {
		group := entry.GetAttributeValue("cn")
		groups = append(groups, group)
	}

	c.log.Info("Queried groups", "count", len(groups))
	return groups, nil
}
