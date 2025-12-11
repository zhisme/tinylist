package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/zhisme/tinylist/internal/config"
)

// BasicAuth returns a middleware that validates Basic Authentication credentials.
func BasicAuth(auth config.AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if !ok {
				unauthorized(w)
				return
			}

			// Constant-time comparison to prevent timing attacks
			usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(auth.Username)) == 1
			passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(auth.Password)) == 1

			if !usernameMatch || !passwordMatch {
				unauthorized(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="TinyList Admin"`)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"Unauthorized"}`))
}
