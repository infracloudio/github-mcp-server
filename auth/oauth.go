package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OAuthConfig struct {
	Provider      *oidc.Provider
	Config        *oauth2.Config
	Verifier      *oidc.IDTokenVerifier
	KeycloakURL   string
	KeycloakRealm string
}

type UserInfo struct {
	Sub               string   `json:"sub"`
	PreferredUsername string   `json:"preferred_username"`
	Email             string   `json:"email"`
	Roles             []string `json:"roles"`
}

func NewOAuthConfig(ctx context.Context) (*OAuthConfig, error) {
	keycloakURL := os.Getenv("KEYCLOAK_URL")
	if keycloakURL == "" {
		keycloakURL = "http://localhost:8080"
	}

	realm := os.Getenv("KEYCLOAK_REALM")
	if realm == "" {
		realm = "mcp-realm"
	}

	provider, err := oidc.NewProvider(ctx, fmt.Sprintf("%s/realms/%s", keycloakURL, realm))
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %v", err)
	}

	config := &oauth2.Config{
		ClientID:     os.Getenv("OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8081/callback",
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "roles"},
	}

	return &OAuthConfig{
		Provider:      provider,
		Config:        config,
		Verifier:      provider.Verifier(&oidc.Config{ClientID: config.ClientID}),
		KeycloakURL:   keycloakURL,
		KeycloakRealm: realm,
	}, nil
}

func (o *OAuthConfig) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	userInfo := &UserInfo{}

	// Get user info from Keycloak
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/realms/%s/protocol/openid-connect/userinfo", o.KeycloakURL, o.KeycloakRealm), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %v", err)
	}

	return userInfo, nil
}
