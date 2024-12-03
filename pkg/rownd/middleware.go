package rownd

import (
	"net/http"
	"strings"
)

// Middleware creates a new HTTP middleware handler for validating Rownd tokens
func (c *Client) Middleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, "no token provided", http.StatusUnauthorized)
				return
			}

			_, token, ok := strings.Cut(header, "Bearer ")
			if !ok {
				http.Error(w, "invalid token format", http.StatusUnauthorized)
				return
			}

			// Validate token
			validToken, err := c.ValidateToken(r.Context(), token)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			// Add token to context
			ctx := AddTokenToCtx(r.Context(), validToken)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
} 