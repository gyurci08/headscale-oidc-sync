package ldap

import (
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/config"
	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/logger"
)

// LDAPClient provides methods to query users and groups.
type LDAPClient interface {
	QueryUsers() ([]User, error)
	QueryGroups() ([]Group, error)
	QueryUsersWithGroups() ([]User, error)
	Close()
}

// Client implements LDAPClient using github.com/go-ldap/ldap.
type Client struct {
	conn   *ldap.Conn
	config config.LdapConfig
	log    logger.ILogger

	// LDAP attribute names for flexibility
	AttrUID      string
	AttrUsername string
	AttrEmail    string
	AttrGroups   string
}

// NewClient creates a new LDAP client connection with config and logger.
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

	client := &Client{
		conn:         conn,
		config:       cfg,
		log:          log,
		AttrUID:      cfg.AttrUID,
		AttrUsername: cfg.AttrUsername,
		AttrEmail:    cfg.AttrEmail,
		AttrGroups:   cfg.AttrGroups,
	}

	log.Debug("Connected to LDAP")
	return client, nil
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

// Close closes the LDAP connection.
func (c *Client) Close() {
	c.conn.Close()
	c.log.Debug("LDAP connection closed")
}

// QueryUsers queries basic users with UID, username, email and all other attributes.
func (c *Client) QueryUsers() ([]User, error) {
	attrs := []string{c.AttrUID, c.AttrUsername, c.AttrEmail} // minimal, add more if needed
	entries, err := c.searchEntries(c.config.UserFilter, attrs)
	if err != nil {
		return nil, err
	}

	users := make([]User, 0, len(entries))
	for _, entry := range entries {
		attrMap := make(map[string]string)
		for _, attr := range entry.Attributes {
			if len(attr.Values) > 0 {
				attrMap[attr.Name] = attr.Values[0]
			}
		}

		user := User{
			UID:        entry.GetAttributeValue(c.AttrUID),
			Username:   entry.GetAttributeValue(c.AttrUsername),
			Email:      entry.GetAttributeValue(c.AttrEmail),
			Attributes: attrMap,
		}
		users = append(users, user)
	}

	c.log.Debug("Queried users", "count", len(users))
	return users, nil
}

// QueryGroups queries all groups with full details matching the group filter.
func (c *Client) QueryGroups() ([]Group, error) {
	attrs := []string{"cn", "description", "member", "distinguishedName", "owner", "manager", "memberOf", "whenCreated", "whenChanged", "displayName", "info"}
	entries, err := c.searchEntries(c.config.GroupFilter, attrs)
	if err != nil {
		return nil, err
	}

	groups := make([]Group, 0, len(entries))
	for _, entry := range entries {
		attrMap := make(map[string]string)
		for _, attr := range entry.Attributes {
			if len(attr.Values) > 0 {
				attrMap[attr.Name] = attr.Values[0]
			}
		}

		group := Group{
			Name:        entry.GetAttributeValue("cn"),
			DN:          entry.DN,
			Description: entry.GetAttributeValue("description"),
			Members:     entry.GetAttributeValues("member"),
			Owner:       entry.GetAttributeValue("owner"),
			Manager:     entry.GetAttributeValue("manager"),
			MemberOf:    entry.GetAttributeValues("memberOf"),
			WhenCreated: entry.GetAttributeValue("whenCreated"),
			WhenChanged: entry.GetAttributeValue("whenChanged"),
			DisplayName: entry.GetAttributeValue("displayName"),
			Info:        entry.GetAttributeValue("info"),
			Attributes:  attrMap,
		}
		groups = append(groups, group)
	}
	return groups, nil
}

// QueryUsersWithGroups queries users along with their group memberships.
func (c *Client) QueryUsersWithGroups() ([]User, error) {
	attrs := []string{c.AttrUID, c.AttrUsername, c.AttrEmail, c.AttrGroups}
	entries, err := c.searchEntries(c.config.UserFilter, attrs)
	if err != nil {
		return nil, err
	}

	users := make([]User, 0, len(entries))
	for _, entry := range entries {
		attrMap := make(map[string]string)
		for _, attr := range entry.Attributes {
			if len(attr.Values) > 0 {
				attrMap[attr.Name] = attr.Values[0]
			}
		}

		var userGroups []Group
		for _, dn := range entry.GetAttributeValues(c.AttrGroups) {
			cn := extractCN(dn)
			if cn != "" {
				userGroups = append(userGroups, Group{Name: cn, DN: dn})
			}
		}

		user := User{
			UID:        entry.GetAttributeValue(c.AttrUID),
			Username:   entry.GetAttributeValue(c.AttrUsername),
			Email:      entry.GetAttributeValue(c.AttrEmail),
			Groups:     userGroups,
			Attributes: attrMap,
		}
		users = append(users, user)
	}

	c.log.Debug("Queried users with groups", "count", len(users))
	return users, nil
}

// queryAttribute returns the values of a single attribute filtered by LDAP filter.
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

// searchEntries executes an LDAP search for the given filter and return specified attributes.
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

// extractGroupNames parses CN from DNs like "CN=Group1,OU=..."
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
		p := strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToLower(p), "cn=") {
			return p[3:]
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
