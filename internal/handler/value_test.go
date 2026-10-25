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

func TestValueHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name      string
		initState storage
		path      string
		code      int
		want      string
	}{
		{
			name:      "positive #1",
			initState: storage{gauges: map[string]float64{"Alloc": 17.3}, counters: map[string]int64{}},
			path:      "/value/gauge/Alloc",
			code:      200,
			want:      "17.3",
		},
		{
			name:      "positive #2",
			initState: storage{gauges: map[string]float64{}, counters: map[string]int64{"PollCount": 42}},
			path:      "/value/counter/PollCount",
			code:      200,
			want:      "42",
		},
		{
			name:      "returns metric of requested type",
			initState: storage{gauges: map[string]float64{"PollCount": 17}, counters: map[string]int64{"PollCount": 42}},
			path:      "/value/counter/PollCount",
			code:      200,
			want:      "42",
		},
		{
			name:      "invalid type",
			initState: storage{gauges: map[string]float64{}, counters: map[string]int64{}},
			path:      "/value/bool/Valid",
			code:      400,
		},
		{
			name:      "unknown metric",
			initState: storage{gauges: map[string]float64{}, counters: map[string]int64{"PollCount": 42}},
			path:      "/value/gauge/PollCount",
			code:      404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)

			w := httptest.NewRecorder()
			h := NewValueHandler(tt.initState)

			r := chi.NewRouter()
			r.Get("/value/{type}/{name}", h.ServeHTTP)
			r.ServeHTTP(w, request)

			result := w.Result()

			defer result.Body.Close()
			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.code, result.StatusCode)
			if len(tt.want) > 0 {
				assert.Equal(t, tt.want, string(body))
			}
		})
	}
}
