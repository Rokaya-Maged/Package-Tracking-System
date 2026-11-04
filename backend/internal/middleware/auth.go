package middleware

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID   int64
	Role string
}

type userContextKey struct{}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		header := r.Header.Get("Authorization")

		if !strings.HasPrefix(header, "Bearer ") {
			http.Error(w, "Login required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")
		secret := os.Getenv("JWT_SECRET")

		if len(secret) < 32 {
			http.Error(w, "Server authentication is not configured", http.StatusInternalServerError)
			return
		}

		token, err := jwt.Parse(
			tokenString,
			func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			},
			jwt.WithValidMethods([]string{"HS256"}),
			jwt.WithExpirationRequired(),
		)

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		idText, ok := claims["sub"].(string)
		if !ok {
			http.Error(w, "Invalid user ID", http.StatusUnauthorized)
			return
		}

		id, err := strconv.ParseInt(idText, 10, 64)
		if err != nil || id <= 0 {
			http.Error(w, "Invalid user ID", http.StatusUnauthorized)
			return
		}

		role, ok := claims["role"].(string)
		if !ok {
			http.Error(w, "Invalid user role", http.StatusUnauthorized)
			return
		}

		user := User{ID: id, Role: role}

		ctx := context.WithValue(r.Context(), userContextKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUser(r *http.Request) (User, bool) {
	user, ok := r.Context().Value(userContextKey{}).(User)
	return user, ok
}
