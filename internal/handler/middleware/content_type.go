package middleware

import "net/http"

func RequireContentType(contentType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Content-Type") != contentType {
				http.Error(w, "Неверный Content-Type", http.StatusBadRequest)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
