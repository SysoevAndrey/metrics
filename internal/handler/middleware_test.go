package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithPathValidation(t *testing.T) {
	tests := []struct {
		name string
		path string
		code int
	}{
		{name: "returns next", path: "/update/counter/PollCount/42", code: 200},
		{name: "throws error", path: "/update/counter//42", code: 404},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.path, nil)

			w := httptest.NewRecorder()
			count, h := newHandler()

			wrapped := WithPathValidation(h)
			assert.Equal(t, 0, *count)

			wrapped.ServeHTTP(w, request)
			result := w.Result()

			if tt.code == http.StatusNotFound {
				assert.Equal(t, http.StatusNotFound, result.StatusCode)
				assert.Equal(t, 0, *count)
			} else {
				assert.Equal(t, 1, *count)
			}
		})
	}
}

func newHandler() (*int, http.HandlerFunc) {
	callCount := 0
	return &callCount, func(w http.ResponseWriter, r *http.Request) {
		callCount++
	}
}
