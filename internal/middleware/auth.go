package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Michealshodipo56/jarcade-backend/internal/auth"
)

type contextKey string

const UserClaimsKey contextKey = "userClaims"

func Auth(tokens *auth.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Authorization required."})
				return
			}

			token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
			claims, err := tokens.Verify(token)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid or expired token."})
				return
			}

			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ClaimsFromContext(ctx context.Context) (auth.Claims, bool) {
	claims, ok := ctx.Value(UserClaimsKey).(auth.Claims)
	return claims, ok
}
