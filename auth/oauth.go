package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

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

type ResourceAccess struct {
	MCPClient struct {
		Roles []string `json:"roles"`
	} `json:"mcp-client"`
}

type UserInfo struct {
	Sub               string         `json:"sub"`
	PreferredUsername string         `json:"preferred_username"`
	Email             string         `json:"email"`
	ResourceAccess    ResourceAccess `json:"resource_access"`
	Roles             []string       `json:"roles,omitempty"`
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
	// Parse the JWT token directly first
	parts := strings.Split(token.AccessToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Add padding if needed
	if l := len(parts[1]) % 4; l > 0 {
		parts[1] += strings.Repeat("=", 4-l)
	}

	// Decode the payload (second part)
	payload, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode token payload: %v", err)
	}

	// Parse into UserInfo struct
	userInfo := &UserInfo{}
	if err := json.Unmarshal(payload, userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse token payload: %v", err)
	}

	// Extract roles from ResourceAccess
	if userInfo.ResourceAccess.MCPClient.Roles != nil && len(userInfo.ResourceAccess.MCPClient.Roles) > 0 {
		userInfo.Roles = userInfo.ResourceAccess.MCPClient.Roles
	} else {
		// Fallback to userinfo endpoint if roles not in token
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

		// Extract roles from ResourceAccess again (in case they were in userinfo)
		if userInfo.ResourceAccess.MCPClient.Roles != nil && len(userInfo.ResourceAccess.MCPClient.Roles) > 0 {
			userInfo.Roles = userInfo.ResourceAccess.MCPClient.Roles
		}
	}

	// Ensure we have roles
	if len(userInfo.Roles) == 0 {
		return nil, fmt.Errorf("no roles found in token or userinfo")
	}

	return userInfo, nil
}
