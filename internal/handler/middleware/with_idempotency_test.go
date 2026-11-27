package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestWithIdempotency(t *testing.T) {
	tests := []struct {
		name         string
		requestCount int
		key          string
		status       int
		wantCalls    int
	}{
		{
			name:         "doesn't call handler twice",
			requestCount: 2,
			key:          "test-key",
			status:       201,
			wantCalls:    1,
		},
		{
			name:         "calls handler on each request if np key",
			requestCount: 2,
			status:       200,
			wantCalls:    2,
		},
		{
			name:         "calls handler on 500 error",
			requestCount: 2,
			key:          "test-key",
			status:       500,
			wantCalls:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			h := func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.WriteHeader(tt.status)
			}

			r := chi.NewRouter()
			r.Use(WithIdempotency(t.Context()))
			r.Post("/", h)

			for i := 0; i < tt.requestCount; i++ {
				request := httptest.NewRequest(http.MethodPost, "/", nil)
				if len(tt.key) > 0 {
					request.Header.Set("Idempotency-Key", tt.key)
				}

				w := httptest.NewRecorder()

				r.ServeHTTP(w, request)
			}

			assert.Equal(t, tt.wantCalls, calls)
		})
	}
}
