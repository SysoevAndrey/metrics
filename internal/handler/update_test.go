package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	repositories "github.com/SysoevAndrey/metrics/internal/repository"
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

func TestUpdateHandler_updateMetric(t *testing.T) {
	type params struct {
		metricType     string
		metricName     string
		metricRawValue string
	}

	tests := []struct {
		name    string
		storage MetricStorage
		params  params
		wantErr bool
	}{
		{
			name:    "positive #1",
			storage: repositories.NewMemStorage(),
			params:  params{metricType: "counter", metricName: "PollCount", metricRawValue: "42"},
			wantErr: false,
		},
		{
			name:    "positive #2",
			storage: repositories.NewMemStorage(),
			params:  params{metricType: "gauge", metricName: "Alloc", metricRawValue: "42"},
			wantErr: false,
		},
		{
			name:    "type mismatch #1",
			storage: repositories.NewMemStorage(),
			params:  params{metricType: "counter", metricName: "Alloc", metricRawValue: "42.17"},
			wantErr: true,
		},
		{
			name:    "type mismatch #2",
			storage: repositories.NewMemStorage(),
			params:  params{metricType: "gauge", metricName: "Alloc", metricRawValue: "one"},
			wantErr: true,
		},
		{
			name:    "unknown type",
			storage: repositories.NewMemStorage(),
			params:  params{metricType: "string", metricName: "Alloc", metricRawValue: "hello"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewUpdateHandler(tt.storage)
			err := h.updateMetric(tt.params.metricType, tt.params.metricName, tt.params.metricRawValue)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		contentType string
		code        int
	}{
		{
			name:        "positive",
			path:        "/update/counter/PollCount/42",
			contentType: "text/plain",
			code:        200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.path, nil)
			request.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()
			h := NewUpdateHandler(repositories.NewMemStorage())

			r := chi.NewRouter()
			r.Post("/update/{type}/{name}/{value}", h.ServeHTTP)
			r.ServeHTTP(w, request)

			result := w.Result()
			assert.Equal(t, tt.code, result.StatusCode)
		})
	}
}
