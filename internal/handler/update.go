package handlers

import (
	"fmt"
	"mime"
	"net/http"
	"strconv"

	models "github.com/SysoevAndrey/metrics/internal/model"
	"github.com/go-chi/chi/v5"
)

type MetricStorage interface {
	SetGauge(name string, value float64)
	AddCounter(name string, value int64)

	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
}

type UpdateHandler struct {
	Storage MetricStorage
}

func (h UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if mediaType != "text/plain" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricRawValue := chi.URLParam(r, "value")

	if isEmptyParam(metricType) || isEmptyParam(metricName) || isEmptyParam(metricRawValue) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.updateMetric(metricType, metricName, metricRawValue)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h UpdateHandler) updateMetric(metricType, metricName, metricRawValue string) error {
	switch metricType {
	case models.Counter:
		value, err := strconv.ParseInt(metricRawValue, 10, 64)
		if err != nil {
			return err
		}
		h.Storage.AddCounter(metricName, value)
	case models.Gauge:
		value, err := strconv.ParseFloat(metricRawValue, 64)
		if err != nil {
			return err
		}
		h.Storage.SetGauge(metricName, value)
	default:
		return fmt.Errorf("unknown metric type: %q", metricType)
	}

	return nil
}

func isEmptyParam(p string) bool {
	return len(p) == 0
}
