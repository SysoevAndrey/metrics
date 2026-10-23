package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	repositories "github.com/SysoevAndrey/metrics/internal/repository"
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
			h := UpdateHandler{Storage: tt.storage}
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
	type pathValues struct {
		metricType  string
		metricName  string
		metricValue string
	}

	tests := []struct {
		name        string
		contentType string
		pathValues  pathValues
		code        int
	}{
		{
			name:        "positive",
			contentType: "text/plain",
			pathValues:  pathValues{metricType: "counter", metricName: "PollCount", metricValue: "42"},
			code:        200,
		},
		{
			name:        "invalid Content-Type",
			contentType: "application/json",
			pathValues:  pathValues{metricType: "counter", metricName: "PollCount", metricValue: "42"},
			code:        400,
		},
		{
			name:        "missing metric type",
			contentType: "text/plain",
			pathValues:  pathValues{metricType: "", metricName: "PollCount", metricValue: "42"},
			code:        400,
		},
		{
			name:        "missing metric name",
			contentType: "text/plain",
			pathValues:  pathValues{metricType: "counter", metricName: "", metricValue: "42"},
			code:        400,
		},
		{
			name:        "missing metric value",
			contentType: "text/plain",
			pathValues:  pathValues{metricType: "counter", metricName: "PollCount", metricValue: ""},
			code:        400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", nil)
			request.SetPathValue("type", tt.pathValues.metricType)
			request.SetPathValue("name", tt.pathValues.metricName)
			request.SetPathValue("value", tt.pathValues.metricValue)
			request.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()
			h := UpdateHandler{Storage: repositories.NewMemStorage()}
			h.ServeHTTP(w, request)

			result := w.Result()
			assert.Equal(t, tt.code, result.StatusCode)
		})
	}
}
