package config

type LdapConfig struct {
	Host        string `validate:"required"`
	Port        int    `validate:"omitempty,gt=0"`
	Protocol    string `validate:"omitempty,oneof=plain ssl tls starttls"`
	BindDN      string `validate:"required"`
	BindPW      string `validate:"required"`
	BaseDN      string `validate:"required"`
	GroupFilter string `validate:"omitempty"`
	UserFilter  string `validate:"omitempty"`
}

func NewLdapConfig() LdapConfig {
	return LdapConfig{
		Host:        getEnvValue("LDAP_HOST", "ldap.example.com"),
		Port:        getEnvInt("LDAP_PORT", 389),
		Protocol:    getEnvValue("LDAP_PROTOCOL", "plain"),
		BindDN:      getEnvValue("LDAP_BIND_DN", "cn=admin,dc=example,dc=com"),
		BindPW:      getEnvValue("LDAP_BIND_PW", "password"),
		BaseDN:      getEnvValue("LDAP_BASE_DN", "dc=example,dc=com"),
		GroupFilter: getEnvValue("LDAP_GROUP_FILTER", "(&(objectClass=group))"),
		UserFilter:  getEnvValue("LDAP_USER_FILTER", "(&(objectClass=person))"),
	}
}
