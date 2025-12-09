package config

type LdapConfig struct {
	Host              string `validate:"required"`
	Port              int    `validate:"omitempty,gt=0"`
	Protocol          string `validate:"omitempty,oneof=plain ssl tls starttls"`
	BindDN            string `validate:"required"`
	BindPW            string `validate:"required"`
	BaseDN            string `validate:"required"`
	GroupFilter       string
	UserFilter        string
	AttrUserUID       string
	AttrUserUsername  string
	AttrUserEmail     string
	AttrUserMemberOf  string
	AttrGroupUID      string
	AttrGroupCN       string
	AttrGroupMember   string
	AttrGroupDN       string
	AttrGroupMemberOf string
}

func NewLdapConfig() LdapConfig {
	return LdapConfig{
		Host:              getEnvValue("LDAP_HOST", ""),
		Port:              getEnvInt("LDAP_PORT", 389),
		Protocol:          getEnvValue("LDAP_PROTOCOL", "plain"),
		BindDN:            getEnvValue("LDAP_BIND_DN", ""),
		BindPW:            getEnvValue("LDAP_BIND_PW", ""),
		BaseDN:            getEnvValue("LDAP_BASE_DN", ""),
		GroupFilter:       getEnvValue("LDAP_GROUP_FILTER", "(&(objectClass=group))"),
		UserFilter:        getEnvValue("LDAP_USER_FILTER", "(&(objectClass=person))"),
		AttrUserUID:       getEnvValue("LDAP_ATTR_USER_UID", "uid"),
		AttrUserUsername:  getEnvValue("LDAP_ATTR_USER_USERNAME", "cn"),
		AttrUserEmail:     getEnvValue("LDAP_ATTR_USER_EMAIL", "mail"),
		AttrUserMemberOf:  getEnvValue("LDAP_ATTR_USER_MEMBER_OF", "memberOf"),
		AttrGroupUID:      getEnvValue("LDAP_ATTR_GROUP_UID", "uid"),
		AttrGroupCN:       getEnvValue("LDAP_ATTR_GROUP_CN", "cn"),
		AttrGroupMember:   getEnvValue("LDAP_ATTR_GROUP_MEMBER", "member"),
		AttrGroupDN:       getEnvValue("LDAP_ATTR_GROUP_DN", "distinguishedName"),
		AttrGroupMemberOf: getEnvValue("LDAP_ATTR_GROUP_MEMBER_OF", "memberOf"),
	}
}
