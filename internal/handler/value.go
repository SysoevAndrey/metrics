package handlers

import (
	"net/http"
	"strconv"

	models "github.com/SysoevAndrey/metrics/internal/model"
	"github.com/go-chi/chi/v5"
)

type ValueHandler struct {
	storage MetricStorage
}

func NewValueHandler(storage MetricStorage) *ValueHandler {
	return &ValueHandler{storage: storage}
}

func (h ValueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	switch metricType {
	case models.Counter:
		value, exists := h.storage.GetCounter(metricName)
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.FormatInt(value, 10)))
	case models.Gauge:
		value, exists := h.storage.GetGauge(metricName)
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.FormatFloat(value, 'f', -1, 64)))
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
