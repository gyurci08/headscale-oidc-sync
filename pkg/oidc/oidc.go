package oidc

import (
	"context"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/config"
	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/logger"
)

type OidcClient struct {
	provider *oidc.Provider
	config   oauth2.Config
}

func NewOidcClient(cfg config.Config, log logger.ILogger) (*OidcClient, error) {
	oidcCfg := cfg.Oidc

	provider, err := oidc.NewProvider(context.Background(), oidcCfg.Issuer)
	if err != nil {
		return nil, err
	}

	config := oauth2.Config{
		ClientID:     oidcCfg.ClientId,
		ClientSecret: oidcCfg.ClientSecret,
		Endpoint:     provider.Endpoint(),
		Scopes:       strings.Split(oidcCfg.Scope, " "),
	}

	return &OidcClient{provider: provider, config: config}, nil
}

func (c *OidcClient) ListGroups(ctx context.Context) ([]string, error) {
	token, err := c.config.PasswordCredentialsToken(ctx, "", "")
	if err != nil {
		return nil, err
	}

	userInfo, err := c.provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
	if err != nil {
		return nil, err
	}

	var claims struct {
		Groups []string `json:"groups"`
	}
	if err := userInfo.Claims(&claims); err != nil {
		return nil, err
	}

	return claims.Groups, nil
}
