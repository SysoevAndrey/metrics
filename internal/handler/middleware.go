package handlers

import (
	"net/http"
	"path"
)

func WithPathValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cleaned := path.Clean(r.URL.Path); r.URL.Path != cleaned {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
