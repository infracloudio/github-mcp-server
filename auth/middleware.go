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

type responseWriter struct {
	status int
	body   string
}

func (w *responseWriter) Header() http.Header {
	return http.Header{}
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body = string(b)
	return len(b), nil
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
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
			if w != nil {
				http.Error(w, "No authorization header", http.StatusUnauthorized)
			}
			return
		}

		bearerToken := strings.TrimPrefix(authHeader, "Bearer ")
		if bearerToken == authHeader {
			if w != nil {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			}
			return
		}

		token := &oauth2.Token{
			AccessToken: bearerToken,
		}

		userInfo, err := am.oauth.GetUserInfo(r.Context(), token)
		if err != nil {
			if w != nil {
				http.Error(w, fmt.Sprintf("Failed to get user info: %v", err), http.StatusUnauthorized)
			}
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), UserContextKey, userInfo)
		if w != nil {
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			mockWriter := &responseWriter{}
			next.ServeHTTP(mockWriter, r.WithContext(ctx))
		}
	})
}

// RequirePermission middleware checks if the user has the required permission
func (am *AuthMiddleware) RequirePermission(permission Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userInfo, ok := r.Context().Value(UserContextKey).(*UserInfo)
			if !ok {
				if w != nil {
					http.Error(w, "User not authenticated", http.StatusUnauthorized)
				}
				return
			}

			if !HasPermission(userInfo.Roles, permission) {
				if w != nil {
					http.Error(w, "Permission denied", http.StatusForbidden)
				}
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
