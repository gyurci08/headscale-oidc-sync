package ldap

// Group represents an LDAP group with full details.
type Group struct {
	UID         string
	DN          string
	Name        string
	DisplayName string
	Description string
	Members     []string
	Owner       string
	Manager     string
	MemberOf    []string
	WhenCreated string
	WhenChanged string
	Info        string
	Attributes  map[string]string
}

// GetAttribute returns the value of an arbitrary attribute if it exists.
func (g *Group) GetAttribute(attr string) string {
	if val, ok := g.Attributes[attr]; ok {
		return val
	}
	return ""
}
