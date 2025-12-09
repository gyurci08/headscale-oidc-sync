package ldap

// User represents an LDAP user with full details.
type User struct {
	UID           string
	DN            string
	Username      string
	Email         string
	DisplayName   string
	Description   string
	FirstName     string
	LastName      string
	HomeDirectory string
	LoginShell    string
	Manager       string
	MemberOf      []string
	WhenChanged   string
	Info          string
	Groups        []Group
	Attributes    map[string]string
}

// GetAttribute returns the value of an arbitrary attribute if it exists.
func (u *User) GetAttribute(attr string) string {
	if val, ok := u.Attributes[attr]; ok {
		return val
	}
	return ""
}
