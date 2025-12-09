package ldap

// User represents an LDAP user with full details.
type User struct {
	UID           string            // unique identifier (e.g., entryUUID, objectGUID)
	DN            string            // distinguishedName
	Username      string            // sAMAccountName or uid
	Email         string            // mail
	DisplayName   string            // displayName
	Description   string            // description
	FirstName     string            // givenName
	LastName      string            // sn
	HomeDirectory string            // homeDirectory
	LoginShell    string            // loginShell
	Manager       string            // manager
	MemberOf      []string          // memberOf (groups the user is a member of)
	WhenCreated   string            // whenCreated
	WhenChanged   string            // whenChanged
	Info          string            // info
	Groups        []Group           // group memberships
	Attributes    map[string]string // all other attributes from LDAP entry
}

// GetAttribute returns the value of an arbitrary attribute if it exists.
func (u *User) GetAttribute(attr string) string {
	if val, ok := u.Attributes[attr]; ok {
		return val
	}
	return ""
}
