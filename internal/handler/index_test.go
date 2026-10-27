package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexHandler_ServeHTTP(t *testing.T) {
	type want struct {
		contentType string
		code        int
		contains    []string
		notContains []string
	}

	tests := []struct {
		name      string
		initState *storage
		want      want
	}{
		{
			name:      "positive",
			initState: &storage{gauges: map[string]float64{"Alloc": 17.3}, counters: map[string]int64{"PollCount": 42}},
			want: want{
				contentType: "text/html",
				code:        200,
				contains:    []string{"Alloc", "17.3", "PollCount", "42"},
				notContains: []string{"No metrics"},
			},
		},
		{
			name:      "no metrics",
			initState: newStorage(),
			want: want{
				contentType: "text/html",
				code:        200,
				contains:    []string{"No metrics"},
				notContains: []string{"Alloc", "17.3", "PollCount", "42"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/", nil)

			w := httptest.NewRecorder()
			h, err := NewIndexHandler(tt.initState)
			require.NoError(t, err)

			r := chi.NewRouter()
			r.Get("/", h.ServeHTTP)
			r.ServeHTTP(w, request)

			result := w.Result()

			defer result.Body.Close()
			bodyRaw, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			body := string(bodyRaw)

			assert.Equal(t, tt.want.code, result.StatusCode)
			assert.Contains(t, result.Header.Get("Content-Type"), tt.want.contentType)

			for _, text := range tt.want.contains {
				assert.Contains(t, body, text)
			}

			for _, text := range tt.want.notContains {
				assert.NotContains(t, body, text)
			}
		})
	}
}
