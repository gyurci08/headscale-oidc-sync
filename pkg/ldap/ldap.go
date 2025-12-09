package ldap

import (
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/config"
	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/logger"
	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/models"
)

type Client struct {
	conn   *ldap.Conn
	config config.LdapConfig
	log    logger.ILogger
}

func NewClient(cfg config.LdapConfig, log logger.ILogger) (*Client, error) {
	url := buildLDAPURL(cfg)
	conn, err := ldap.DialURL(url)
	if err != nil {
		log.Error("Failed to connect to LDAP", "error", err)
		return nil, err
	}

	if strings.EqualFold(cfg.Protocol, "starttls") {
		if err = conn.StartTLS(nil); err != nil {
			log.Error("Failed to start TLS", "error", err)
			conn.Close()
			return nil, err
		}
	}

	if err = conn.Bind(cfg.BindDN, cfg.BindPW); err != nil {
		log.Error("Failed to bind to LDAP", "error", err)
		conn.Close()
		return nil, err
	}

	log.Debug("Connected to LDAP")
	return &Client{conn: conn, config: cfg, log: log}, nil
}

func buildLDAPURL(cfg config.LdapConfig) string {
	switch strings.ToLower(cfg.Protocol) {
	case "ssl", "tls":
		return fmt.Sprintf("ldaps://%s:%d", cfg.Host, cfg.Port)
	case "starttls":
		return fmt.Sprintf("ldap://%s:%d", cfg.Host, cfg.Port)
	default:
		return fmt.Sprintf("ldap://%s:%d", cfg.Host, cfg.Port)
	}
}

func (c *Client) Close() {
	c.conn.Close()
	c.log.Debug("LDAP connection closed")
}

func (c *Client) QueryUsers() ([]models.User, error) {
	return c.QueryUsersWithRoles()
}

func (c *Client) QueryGroups() ([]string, error) {
	return c.queryAttribute("cn", c.config.GroupFilter)
}

func (c *Client) QueryUsersWithRoles() ([]models.User, error) {
	entries, err := c.searchEntries(c.config.UserFilter, []string{"mail", "uid", "cn", "memberOf"})
	if err != nil {
		return nil, err
	}

	users := make([]models.User, 0, len(entries))
	for _, entry := range entries {
		user := models.User{
			Email:    entry.GetAttributeValue("mail"),
			Username: firstNonEmpty(entry.GetAttributeValue("uid"), entry.GetAttributeValue("cn")),
			Roles:    extractGroupNames(entry.GetAttributeValues("memberOf")),
		}
		if user.Email != "" {
			users = append(users, user)
		}
	}
	c.log.Debug("Queried users with roles", "count", len(users))
	return users, nil
}

func (c *Client) queryAttribute(attr, filter string) ([]string, error) {
	entries, err := c.searchEntries(filter, []string{attr})
	if err != nil {
		return nil, err
	}

	values := make([]string, 0, len(entries))
	for _, entry := range entries {
		if val := entry.GetAttributeValue(attr); val != "" {
			values = append(values, val)
		}
	}
	c.log.Debug(fmt.Sprintf("Queried %s", attr), "count", len(values))
	return values, nil
}

func (c *Client) searchEntries(filter string, attrs []string) ([]*ldap.Entry, error) {
	filter = wrapFilter(filter)
	c.log.Debug("LDAP search filter", "filter", filter)
	c.log.Debug("LDAP search baseDN", "baseDN", c.config.BaseDN)

	sr, err := c.conn.Search(ldap.NewSearchRequest(
		c.config.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		attrs,
		nil,
	))
	if err != nil {
		c.log.Error("LDAP search failed", "error", err)
		return nil, err
	}
	return sr.Entries, nil
}

func extractGroupNames(dns []string) []string {
	groups := make([]string, 0, len(dns))
	for _, dn := range dns {
		if cn := extractCN(dn); cn != "" {
			groups = append(groups, cn)
		}
	}
	return groups
}

func extractCN(dn string) string {
	for _, part := range strings.Split(dn, ",") {
		if strings.HasPrefix(strings.ToLower(part), "cn=") {
			return part[3:]
		}
	}
	return ""
}

func wrapFilter(filter string) string {
	if filter == "" || strings.HasPrefix(filter, "(&") {
		return filter
	}
	return fmt.Sprintf("(&%s)", filter)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
