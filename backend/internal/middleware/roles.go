package middleware

import "net/http"

func RequireRole(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user, ok := GetUser(r)

		if !ok {
			http.Error(w, "Login required", http.StatusUnauthorized)
			return
		}

		if user.Role != role {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
