package middleware

import (
	"context"
	"net/http"
	"strings"

	"profit-ecommerce/internal/auth"
)

type contextKey string

const UserIDKey contextKey = "userID"

// RequireAuth es un middleware HTTP que exige un Bearer token.
// Valida el token contra AuthService e inyecta el identificador en el Context actual en caso de éxito.
func RequireAuth(authSvc auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			if token == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := authSvc.ValidateToken(r.Context(), token)
			if err != nil {
				// Cualquier fallo en la autorización retorna un genérico sin revelar la excepción interna.
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Inyección segura del UserID extraído sin requerir re-conexiones por cada endpoint a posteriori.
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extrae la const asignada del UserID del context HTTP original.
func GetUserID(ctx context.Context) string {
	val, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return ""
	}
	return val
}
