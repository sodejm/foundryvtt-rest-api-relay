package middleware

import "net/http"

func RequireAdmin() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler { return h }
}
func Other() func(http.Handler) http.Handler { return func(h http.Handler) http.Handler { return h } }
