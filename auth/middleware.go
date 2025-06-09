package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

type AuthMiddleware struct {
	oauth *OAuthConfig
}

func NewAuthMiddleware(oauth *OAuthConfig) *AuthMiddleware {
	return &AuthMiddleware{
		oauth: oauth,
	}
}

// Authenticate middleware verifies the token and adds user info to the context
func (am *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "No authorization header", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.TrimPrefix(authHeader, "Bearer ")
		if bearerToken == authHeader {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := &oauth2.Token{
			AccessToken: bearerToken,
		}

		userInfo, err := am.oauth.GetUserInfo(r.Context(), token)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get user info: %v", err), http.StatusUnauthorized)
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), UserContextKey, userInfo)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermission middleware checks if the user has the required permission
func (am *AuthMiddleware) RequirePermission(permission Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userInfo, ok := r.Context().Value(UserContextKey).(*UserInfo)
			if !ok {
				http.Error(w, "User not authenticated", http.StatusUnauthorized)
				return
			}

			if !HasPermission(userInfo.Roles, permission) {
				http.Error(w, "Permission denied", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext retrieves the user info from the context
func GetUserFromContext(ctx context.Context) (*UserInfo, error) {
	user, ok := ctx.Value(UserContextKey).(*UserInfo)
	if !ok {
		return nil, fmt.Errorf("user not found in context")
	}
	return user, nil
}
