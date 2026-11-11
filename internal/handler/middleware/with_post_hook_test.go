package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithPostHook(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantCalls int
	}{
		{
			name:      "calls on 200",
			path:      "/success",
			wantCalls: 1,
		},
		{
			name:      "skips on 500",
			path:      "/error",
			wantCalls: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			fn := newFn(&calls)

			request, err := http.NewRequest(http.MethodGet, tt.path, nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Use(WithPostHook(fn))
			r.Get("/success", newHandlerFunc(http.StatusOK))
			r.Get("/error", newHandlerFunc(http.StatusInternalServerError))
			r.ServeHTTP(w, request)

			assert.Equal(t, tt.wantCalls, calls)
		})
	}
}

func newFn(calls *int) func() error {
	return func() error {
		*calls++
		return nil
	}
}

func newHandlerFunc(statusCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}
}
