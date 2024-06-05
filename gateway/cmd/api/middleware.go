package main

import (
	"net/http"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		// Assuming validateToken is a function that validates your token
		if !validateToken(token) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Call the next handler if the token is valid
		next.ServeHTTP(w, r)
	})
}

func validateToken(token string) bool {
	return token == "valid-token"
}
