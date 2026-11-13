package middleware

import (
	"net/http"

	"github.com/SysoevAndrey/metrics/internal/logger"
	"go.uber.org/zap"
)

type postHookResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *postHookResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func WithPostHook(fn func() error) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writer := postHookResponseWriter{ResponseWriter: w}

			h.ServeHTTP(&writer, r)

			if writer.statusCode < 300 {
				if err := fn(); err != nil {
					logger.Log.Error("failed to execute post hook", zap.Error(err))
				}
			}
		})
	}
}
