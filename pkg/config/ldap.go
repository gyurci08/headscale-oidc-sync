package config

type LdapConfig struct {
	Host         string `validate:"required"`
	Port         int    `validate:"omitempty,gt=0"`
	Protocol     string `validate:"omitempty,oneof=plain ssl tls starttls"`
	BindDN       string `validate:"required"`
	BindPW       string `validate:"required"`
	BaseDN       string `validate:"required"`
	GroupFilter  string `validate:"omitempty"`
	UserFilter   string `validate:"omitempty"`
	AttrUID      string `validate:"omitempty"`
	AttrUsername string `validate:"omitempty"`
	AttrEmail    string `validate:"omitempty"`
	AttrGroups   string `validate:"omitempty"`
}

func NewLdapConfig() LdapConfig {
	return LdapConfig{
		Host:         getEnvValue("LDAP_HOST", ""),
		Port:         getEnvInt("LDAP_PORT", 389),
		Protocol:     getEnvValue("LDAP_PROTOCOL", "plain"),
		BindDN:       getEnvValue("LDAP_BIND_DN", ""),
		BindPW:       getEnvValue("LDAP_BIND_PW", ""),
		BaseDN:       getEnvValue("LDAP_BASE_DN", ""),
		GroupFilter:  getEnvValue("LDAP_GROUP_FILTER", "(&(objectClass=group))"),
		UserFilter:   getEnvValue("LDAP_USER_FILTER", "(&(objectClass=person))"),
		AttrUID:      getEnvValue("LDAP_ATTR_UID", "uid"),
		AttrUsername: getEnvValue("LDAP_ATTR_USERNAME", "cn"),
		AttrEmail:    getEnvValue("LDAP_ATTR_EMAIL", "mail"),
		AttrGroups:   getEnvValue("LDAP_ATTR_GROUPS", "memberOf"),
	}
}
