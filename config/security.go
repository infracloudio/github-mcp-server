package config

import (
	"os"
)

// SecurityConfig holds all security-related configuration
type SecurityConfig struct {
	KeycloakURL       string
	KeycloakRealm     string
	OAuthClientID     string
	OAuthClientSecret string
	ServerPort        string
	AllowedOrigins    []string
}

// NewSecurityConfig creates a new security configuration from environment variables
func NewSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		KeycloakURL:       getEnvOrDefault("KEYCLOAK_URL", "http://localhost:8080"),
		KeycloakRealm:     getEnvOrDefault("KEYCLOAK_REALM", "mcp-realm"),
		OAuthClientID:     getEnvOrDefault("OAUTH_CLIENT_ID", ""),
		OAuthClientSecret: getEnvOrDefault("OAUTH_CLIENT_SECRET", ""),
		ServerPort:        getEnvOrDefault("SERVER_PORT", "8081"),
		AllowedOrigins:    []string{"http://localhost:8081"},
	}
}

// Validate checks if all required configuration is present
func (c *SecurityConfig) Validate() error {
	if c.OAuthClientID == "" {
		return ErrMissingClientID
	}
	if c.OAuthClientSecret == "" {
		return ErrMissingClientSecret
	}
	return nil
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Error definitions
var (
	ErrMissingClientID     = &ConfigError{msg: "OAuth client ID is required"}
	ErrMissingClientSecret = &ConfigError{msg: "OAuth client secret is required"}
)

// ConfigError represents a configuration error
type ConfigError struct {
	msg string
}

func (e *ConfigError) Error() string {
	return e.msg
}
