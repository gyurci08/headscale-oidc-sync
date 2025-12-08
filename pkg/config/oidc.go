package config

type OidcConfig struct {
	Issuer       string `validate:"required,url"`
	ClientId     string `validate:"required"`
	ClientSecret string `validate:"required"`
	Scope        string `validate:"omitempty"`
}

func NewOidcConfig() OidcConfig {
	return OidcConfig{
		Issuer:       getEnv("OIDC_ISSUER", "https://oidc.example.com/application/o/app/"),
		ClientId:     getEnv("OIDC_CLIENT_ID", "id"),
		ClientSecret: getEnv("OIDC_CLIENT_SECRET", "secret"),
		Scope:        getEnv("OIDC_SCOPE", "openid email profile"),
	}
}
