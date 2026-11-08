package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	models "github.com/SysoevAndrey/metrics/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestIsEmptyParam(t *testing.T) {
	tests := []struct {
		name  string
		param string
		want  bool
	}{
		{name: "complete", param: "42", want: false},
		{name: "empty", param: "", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isEmptyParam(tt.param))
		})
	}
}

func TestUpdateHandler_update(t *testing.T) {
	tests := []struct {
		name    string
		storage MetricStorage
		metrics models.Metrics
		wantErr bool
	}{
		{
			name:    "positive #1",
			storage: newStorage(),
			metrics: models.Metrics{ID: "PollCount", MType: "counter", Delta: new(int64(42))},
			wantErr: false,
		},
		{
			name:    "positive #2",
			storage: newStorage(),
			metrics: models.Metrics{ID: "Alloc", MType: "gauge", Value: new(42.17)},
			wantErr: false,
		},
		{
			name:    "type mismatch #1",
			storage: newStorage(),
			metrics: models.Metrics{ID: "Alloc", MType: "counter", Value: new(42.17)},
			wantErr: true,
		},
		{
			name:    "unknown type",
			storage: newStorage(),
			metrics: models.Metrics{ID: "Alloc", MType: "string"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewUpdateHandler(tt.storage)
			err := h.update(tt.metrics)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_newMetrics(t *testing.T) {
	type params struct {
		id       string
		mType    string
		rawValue string
	}

	type want struct {
		metrics *models.Metrics
		err     error
	}

	tests := []struct {
		name   string
		params params
		want   want
	}{
		{
			name:   "creates counter",
			params: params{id: "PollCount", mType: "counter", rawValue: "42"},
			want:   want{metrics: &models.Metrics{ID: "PollCount", MType: "counter", Delta: new(int64(42))}},
		},
		{
			name:   "creates gauge",
			params: params{id: "Alloc", mType: "gauge", rawValue: "17.03"},
			want:   want{metrics: &models.Metrics{ID: "Alloc", MType: "gauge", Value: new(17.03)}},
		},
		{
			name:   "throws on invalid value #1",
			params: params{id: "Alloc", mType: "gauge", rawValue: "abc"},
			want:   want{err: ErrParseValue},
		},
		{
			name:   "throws on invalid value #2",
			params: params{id: "PollCount", mType: "counter", rawValue: "17.03"},
			want:   want{err: ErrParseValue},
		},
		{
			name:   "throws on invalid type",
			params: params{id: "Valid", mType: "bool", rawValue: "true"},
			want:   want{err: ErrUnknownType},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics, err := newMetrics(tt.params.id, tt.params.mType, tt.params.rawValue)

			assert.ErrorIs(t, err, tt.want.err)
			assert.Equal(t, tt.want.metrics, metrics)
		})
	}
}

func TestUpdateHandler_ByPath(t *testing.T) {
	type want struct {
		code     int
		gauges   map[string]float64
		counters map[string]int64
	}

	tests := []struct {
		name        string
		path        string
		contentType string
		want        want
	}{
		{
			name:        "add counter",
			path:        "/update/counter/PollCount/42",
			contentType: "text/plain",
			want: want{
				code:     200,
				gauges:   map[string]float64{},
				counters: map[string]int64{"PollCount": 42},
			},
		},
		{
			name:        "set gauge",
			path:        "/update/gauge/Alloc/17.03",
			contentType: "text/plain",
			want: want{
				code:     200,
				gauges:   map[string]float64{"Alloc": 17.03},
				counters: map[string]int64{},
			},
		},
		{
			name:        "bad request",
			path:        "/update/counter/Alloc/abc",
			contentType: "text/plain",
			want: want{
				code:     400,
				gauges:   map[string]float64{},
				counters: map[string]int64{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.path, nil)
			request.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()
			s := newStorage()
			h := NewUpdateHandler(s)

			r := chi.NewRouter()
			r.Post("/update/{type}/{name}/{value}", h.ByPath)
			r.ServeHTTP(w, request)

			result := w.Result()
			assert.Equal(t, tt.want.code, result.StatusCode)
			assert.EqualValues(t, tt.want.gauges, s.gauges)
			assert.EqualValues(t, tt.want.counters, s.counters)
		})
	}
}

func TestUpdateHandler_ByBody(t *testing.T) {
	type want struct {
		code     int
		gauges   map[string]float64
		counters map[string]int64
	}

	tests := []struct {
		name        string
		body        string
		contentType string
		want        want
	}{
		{
			name:        "add counter",
			body:        `{"id": "PollCount", "type": "counter", "delta": 42}`,
			contentType: "application/json",
			want: want{
				code:     200,
				gauges:   map[string]float64{},
				counters: map[string]int64{"PollCount": 42},
			},
		},
		{
			name:        "set gauge",
			body:        `{"id": "Alloc", "type": "gauge", "value": 17.03}`,
			contentType: "application/json",
			want: want{
				code:     200,
				gauges:   map[string]float64{"Alloc": 17.03},
				counters: map[string]int64{},
			},
		},
		{
			name:        "bad request",
			body:        `{"id": "Alloc", "type": "counter", "delta": "abc"}`,
			contentType: "application/json",
			want: want{
				code:     400,
				gauges:   map[string]float64{},
				counters: map[string]int64{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyReader := bytes.NewReader([]byte(tt.body))
			request := httptest.NewRequest(http.MethodPost, "/update", bodyReader)
			request.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()
			s := newStorage()
			h := NewUpdateHandler(s)

			r := chi.NewRouter()
			r.Post("/update", h.ByBody)
			r.ServeHTTP(w, request)

			result := w.Result()
			assert.Equal(t, tt.want.code, result.StatusCode)
			assert.EqualValues(t, tt.want.gauges, s.gauges)
			assert.EqualValues(t, tt.want.counters, s.counters)
		})
	}
}
