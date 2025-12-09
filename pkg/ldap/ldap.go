package ldap

import (
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/config"
	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/logger"
)

// LDAP attribute constants for users
const (
	UserAttrDN            = "dn"
	UserAttrDisplayName   = "displayName"
	UserAttrDescription   = "description"
	UserAttrGivenName     = "givenName"
	UserAttrSurname       = "sn"
	UserAttrHomeDirectory = "homeDirectory"
	UserAttrLoginShell    = "loginShell"
	UserAttrManager       = "manager"
	UserAttrWhenChanged   = "whenChanged"
	UserAttrInfo          = "info"
)

// LDAP attribute constants for groups
const (
	GroupAttrDescription = "description"
	GroupAttrOwner       = "owner"
	GroupAttrManager     = "manager"
	GroupAttrWhenCreated = "whenCreated"
	GroupAttrWhenChanged = "whenChanged"
	GroupAttrDisplayName = "displayName"
	GroupAttrInfo        = "info"
)

// LDAPClient provides methods to query users and groups.
type LDAPClient interface {
	QueryUsers() ([]User, error)
	QueryGroups() ([]Group, error)
	QueryUsersWithGroups() ([]User, error)
	QueryUser(uid string) (*User, error)
	QueryGroup(uid string) (*Group, error)
	Close()
}

// Client implements LDAPClient using go-ldap.
type Client struct {
	conn   *ldap.Conn
	config config.LdapConfig
	log    logger.ILogger

	// LDAP attribute names for flexibility
	AttrUserUID       string
	AttrUsername      string
	AttrEmail         string
	AttrUserMemberOf  string
	AttrGroupUID      string
	AttrGroupCN       string
	AttrGroupMember   string
	AttrGroupDN       string
	AttrGroupMemberOf string
}

// Common attribute sets reused across queries
var (
	userBaseAttrs = []string{
		UserAttrDN,
		UserAttrDisplayName,
		UserAttrDescription,
		UserAttrGivenName,
		UserAttrSurname,
		UserAttrHomeDirectory,
		UserAttrLoginShell,
		UserAttrManager,
		UserAttrWhenChanged,
		UserAttrInfo,
	}

	groupBaseAttrs = []string{
		GroupAttrDescription,
		GroupAttrOwner,
		GroupAttrManager,
		GroupAttrWhenCreated,
		GroupAttrWhenChanged,
		GroupAttrDisplayName,
		GroupAttrInfo,
	}
)

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
		conn:              conn,
		config:            cfg,
		log:               log,
		AttrUserUID:       cfg.AttrUserUID,
		AttrUsername:      cfg.AttrUserUsername,
		AttrEmail:         cfg.AttrUserEmail,
		AttrUserMemberOf:  cfg.AttrUserMemberOf,
		AttrGroupUID:      cfg.AttrGroupUID,
		AttrGroupCN:       cfg.AttrGroupCN,
		AttrGroupMember:   cfg.AttrGroupMember,
		AttrGroupDN:       cfg.AttrGroupDN,
		AttrGroupMemberOf: cfg.AttrGroupMemberOf,
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

// makeUserAttrs returns the full attribute list to fetch for users, including configurable attrs.
func (c *Client) makeUserAttrs(includeMemberOf bool) []string {
	attrs := []string{
		c.AttrUserUID,
		c.AttrUsername,
		c.AttrEmail,
	}
	attrs = append(attrs, userBaseAttrs...)
	if includeMemberOf {
		attrs = append(attrs, c.AttrUserMemberOf)
	}
	return attrs
}

// makeGroupAttrs returns the full attribute list to fetch for groups.
func (c *Client) makeGroupAttrs() []string {
	attrs := []string{
		c.AttrGroupUID,
		c.AttrGroupCN,
		c.AttrGroupDN,
		c.AttrGroupMember,
		c.AttrGroupMemberOf,
	}
	attrs = append(attrs, groupBaseAttrs...)
	return attrs
}

// mapEntryToUser maps a single LDAP entry to User struct.
func (c *Client) mapEntryToUser(entry *ldap.Entry, includeGroups bool) User {
	attrMap := make(map[string]string)
	for _, attr := range entry.Attributes {
		if len(attr.Values) > 0 {
			attrMap[attr.Name] = attr.Values[0]
		}
	}

	user := User{
		UID:           entry.GetAttributeValue(c.AttrUserUID),
		DN:            entry.DN,
		Username:      entry.GetAttributeValue(c.AttrUsername),
		Email:         entry.GetAttributeValue(c.AttrEmail),
		DisplayName:   entry.GetAttributeValue(UserAttrDisplayName),
		Description:   entry.GetAttributeValue(UserAttrDescription),
		FirstName:     entry.GetAttributeValue(UserAttrGivenName),
		LastName:      entry.GetAttributeValue(UserAttrSurname),
		HomeDirectory: entry.GetAttributeValue(UserAttrHomeDirectory),
		LoginShell:    entry.GetAttributeValue(UserAttrLoginShell),
		Manager:       entry.GetAttributeValue(UserAttrManager),
		MemberOf:      entry.GetAttributeValues(c.AttrUserMemberOf),
		WhenChanged:   entry.GetAttributeValue(UserAttrWhenChanged),
		Info:          entry.GetAttributeValue(UserAttrInfo),
		Attributes:    attrMap,
	}

	if includeGroups {
		var userGroups []Group
		for _, dn := range entry.GetAttributeValues(c.AttrUserMemberOf) {
			cn := extractCN(dn)
			if cn != "" {
				userGroups = append(userGroups, Group{Name: cn, DN: dn})
			}
		}
		user.Groups = userGroups
	}

	return user
}

// mapEntryToGroup maps a single LDAP entry to Group struct.
func (c *Client) mapEntryToGroup(entry *ldap.Entry) Group {
	attrMap := make(map[string]string)
	for _, attr := range entry.Attributes {
		if len(attr.Values) > 0 {
			attrMap[attr.Name] = attr.Values[0]
		}
	}

	return Group{
		UID:         entry.GetAttributeValue(c.AttrGroupUID),
		Name:        entry.GetAttributeValue(c.AttrGroupCN),
		DN:          entry.DN,
		Description: entry.GetAttributeValue(GroupAttrDescription),
		Members:     entry.GetAttributeValues(c.AttrGroupMember),
		Owner:       entry.GetAttributeValue(GroupAttrOwner),
		Manager:     entry.GetAttributeValue(GroupAttrManager),
		MemberOf:    entry.GetAttributeValues(c.AttrGroupMemberOf),
		WhenCreated: entry.GetAttributeValue(GroupAttrWhenCreated),
		WhenChanged: entry.GetAttributeValue(GroupAttrWhenChanged),
		DisplayName: entry.GetAttributeValue(GroupAttrDisplayName),
		Info:        entry.GetAttributeValue(GroupAttrInfo),
		Attributes:  attrMap,
	}
}

// QueryUser queries a single user by UID.
func (c *Client) QueryUser(uid string) (*User, error) {
	filter := fmt.Sprintf("(%s=%s)", c.AttrUserUID, ldap.EscapeFilter(uid))
	entries, err := c.searchEntries(filter, c.makeUserAttrs(true))
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("user not found: %s", uid)
	}

	user := c.mapEntryToUser(entries[0], true)
	c.log.Debug("Queried user", "uid", uid)
	return &user, nil
}

// QueryGroup queries a single group by UID.
func (c *Client) QueryGroup(uid string) (*Group, error) {
	filter := fmt.Sprintf("(%s=%s)", c.AttrGroupUID, ldap.EscapeFilter(uid))
	entries, err := c.searchEntries(filter, c.makeGroupAttrs())
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("group not found: %s", uid)
	}

	group := c.mapEntryToGroup(entries[0])
	c.log.Debug("Queried group", "uid", uid)
	return &group, nil
}

// QueryUsers queries full user details (no embedded groups).
func (c *Client) QueryUsers() ([]User, error) {
	entries, err := c.searchEntries(c.config.UserFilter, c.makeUserAttrs(false))
	if err != nil {
		return nil, err
	}

	users := make([]User, 0, len(entries))
	for _, entry := range entries {
		users = append(users, c.mapEntryToUser(entry, false))
	}

	c.log.Debug("Queried users", "count", len(users))
	return users, nil
}

// QueryGroups queries full group details.
func (c *Client) QueryGroups() ([]Group, error) {
	entries, err := c.searchEntries(c.config.GroupFilter, c.makeGroupAttrs())
	if err != nil {
		return nil, err
	}

	groups := make([]Group, 0, len(entries))
	for _, entry := range entries {
		groups = append(groups, c.mapEntryToGroup(entry))
	}

	c.log.Debug("Queried groups", "count", len(groups))
	return groups, nil
}

// QueryUsersWithGroups queries users including embedded group structs.
func (c *Client) QueryUsersWithGroups() ([]User, error) {
	entries, err := c.searchEntries(c.config.UserFilter, c.makeUserAttrs(true))
	if err != nil {
		return nil, err
	}

	users := make([]User, 0, len(entries))
	for _, entry := range entries {
		users = append(users, c.mapEntryToUser(entry, true))
	}

	c.log.Debug("Queried users with groups", "count", len(users))
	return users, nil
}

// searchEntries executes an LDAP search for the given filter and specified attributes.
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

// extractCN extracts CN from a DN string.
func extractCN(dn string) string {
	for _, part := range strings.Split(dn, ",") {
		p := strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToLower(p), "cn=") {
			return p[3:]
		}
	}
	return ""
}

// wrapFilter ensures the filter is wrapped in an AND clause if needed.
func wrapFilter(filter string) string {
	if filter == "" || strings.HasPrefix(filter, "(&") {
		return filter
	}
	return fmt.Sprintf("(&%s)", filter)
}
