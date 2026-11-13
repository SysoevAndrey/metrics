package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	models "github.com/SysoevAndrey/metrics/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValueHandler_value(t *testing.T) {
	type want struct {
		metrics *models.Metrics
		err     error
	}

	tests := []struct {
		name    string
		storage MetricStorage
		data    models.Metrics
		want    want
	}{
		{
			name:    "positive #1",
			storage: &storage{gauges: map[string]float64{"Alloc": 17.3}, counters: map[string]int64{}},
			data:    models.Metrics{MType: "gauge", ID: "Alloc"},
			want:    want{metrics: &models.Metrics{ID: "Alloc", MType: "gauge", Value: new(17.3)}},
		},
		{
			name:    "positive #2",
			storage: &storage{gauges: map[string]float64{}, counters: map[string]int64{"PollCount": 42}},
			data:    models.Metrics{MType: "counter", ID: "PollCount"},
			want:    want{metrics: &models.Metrics{ID: "PollCount", MType: "counter", Delta: new(int64(42))}},
		},
		{
			name:    "not found #1",
			storage: &storage{gauges: map[string]float64{}, counters: map[string]int64{}},
			data:    models.Metrics{MType: "counter", ID: "PollCount"},
			want:    want{err: ErrNotFound},
		},
		{
			name:    "not found #2",
			storage: &storage{gauges: map[string]float64{}, counters: map[string]int64{}},
			data:    models.Metrics{MType: "gauge", ID: "Alloc"},
			want:    want{err: ErrNotFound},
		},
		{
			name:    "invalid type",
			storage: &storage{gauges: map[string]float64{}, counters: map[string]int64{}},
			data:    models.Metrics{MType: "bool", ID: "Valid"},
			want:    want{err: ErrUnknownType},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewValueHandler(tt.storage)
			metrics, err := h.value(tt.data)

			assert.ErrorIs(t, err, tt.want.err)
			assert.EqualValues(t, tt.want.metrics, metrics)
		})
	}
}

func TestValueHandler_ByPath(t *testing.T) {
	tests := []struct {
		name      string
		initState *storage
		path      string
		code      int
		want      string
	}{
		{
			name:      "positive #1",
			initState: &storage{gauges: map[string]float64{"Alloc": 17.3}, counters: map[string]int64{}},
			path:      "/value/gauge/Alloc",
			code:      200,
			want:      "17.3",
		},
		{
			name:      "positive #2",
			initState: &storage{gauges: map[string]float64{}, counters: map[string]int64{"PollCount": 42}},
			path:      "/value/counter/PollCount",
			code:      200,
			want:      "42",
		},
		{
			name:      "returns metric of requested type",
			initState: &storage{gauges: map[string]float64{"PollCount": 17}, counters: map[string]int64{"PollCount": 42}},
			path:      "/value/counter/PollCount",
			code:      200,
			want:      "42",
		},
		{
			name:      "invalid type",
			initState: &storage{gauges: map[string]float64{}, counters: map[string]int64{}},
			path:      "/value/bool/Valid",
			code:      400,
		},
		{
			name:      "unknown metric",
			initState: &storage{gauges: map[string]float64{}, counters: map[string]int64{"PollCount": 42}},
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
			r.Get("/value/{type}/{name}", h.ByPath)
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

func TestValueHandler_ByBody(t *testing.T) {
	tests := []struct {
		name      string
		initState *storage
		body      string
		code      int
		want      string
	}{
		{
			name:      "positive #1",
			initState: &storage{gauges: map[string]float64{"Alloc": 17.3}, counters: map[string]int64{}},
			body:      `{"id": "Alloc", "type": "gauge"}`,
			code:      200,
			want:      `{"id": "Alloc", "type": "gauge", "value": 17.3}`,
		},
		{
			name:      "positive #2",
			initState: &storage{gauges: map[string]float64{}, counters: map[string]int64{"PollCount": 42}},
			body:      `{"id": "PollCount", "type": "counter"}`,
			code:      200,
			want:      `{"id": "PollCount", "type": "counter", "delta": 42}`,
		},
		{
			name:      "returns metric of requested type",
			initState: &storage{gauges: map[string]float64{"PollCount": 17}, counters: map[string]int64{"PollCount": 42}},
			body:      `{"id": "PollCount", "type": "counter"}`,
			code:      200,
			want:      `{"id": "PollCount", "type": "counter", "delta": 42}`,
		},
		{
			name:      "invalid type",
			initState: &storage{gauges: map[string]float64{}, counters: map[string]int64{}},
			body:      `{"id": "Valid", "type": "bool"}`,
			code:      400,
		},
		{
			name:      "unknown metric",
			initState: &storage{gauges: map[string]float64{}, counters: map[string]int64{"PollCount": 42}},
			body:      `{"id": "PollCount", "type": "gauge"}`,
			code:      404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyReader := bytes.NewReader([]byte(tt.body))
			request := httptest.NewRequest(http.MethodPost, "/value", bodyReader)
			request.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h := NewValueHandler(tt.initState)

			r := chi.NewRouter()
			r.Post("/value", h.ByBody)
			r.ServeHTTP(w, request)

			result := w.Result()

			defer result.Body.Close()
			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.code, result.StatusCode)
			if len(tt.want) > 0 {
				assert.JSONEq(t, tt.want, string(body))
			}
		})
	}
}
