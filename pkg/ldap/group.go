package ldap

// Group represents an LDAP group with full details.
type Group struct {
	Name        string            // cn
	DN          string            // distinguishedName
	Description string            // description
	Members     []string          // member
	Owner       string            // owner
	Manager     string            // manager
	MemberOf    []string          // memberOf (for nested groups)
	WhenCreated string            // whenCreated
	WhenChanged string            // whenChanged
	DisplayName string            // displayName
	Info        string            // info
	Attributes  map[string]string // all other attributes from LDAP entry
}

// GetAttribute returns the value of an arbitrary attribute if it exists.
func (g *Group) GetAttribute(attr string) string {
	if val, ok := g.Attributes[attr]; ok {
		return val
	}
	return ""
}
