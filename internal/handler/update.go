package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/SysoevAndrey/metrics/internal/logger"
	models "github.com/SysoevAndrey/metrics/internal/model"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type UpdateHandler struct {
	storage MetricStorage
}

func NewUpdateHandler(storage MetricStorage) *UpdateHandler {
	return &UpdateHandler{storage: storage}
}

func (h *UpdateHandler) ByPath(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricRawValue := chi.URLParam(r, "value")

	if isEmptyParam(metricType) || isEmptyParam(metricName) || isEmptyParam(metricRawValue) {
		logger.Log.Debug("found empty param")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metrics, err := newMetrics(metricName, metricType, metricRawValue)
	if err != nil {
		logger.Log.Debug(err.Error(), zap.String("id", metricName))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.update(*metrics); err != nil {
		logger.Log.Debug(err.Error(), zap.String("id", metrics.ID))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UpdateHandler) ByBody(w http.ResponseWriter, r *http.Request) {
	var metrics models.Metrics

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&metrics); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.update(metrics); err != nil {
		logger.Log.Debug(err.Error(), zap.String("id", metrics.ID))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UpdateHandler) update(metrics models.Metrics) error {
	switch metrics.MType {
	case models.Counter:
		if metrics.Delta == nil {
			return fmt.Errorf("no counter delta found for: %s", metrics.ID)
		}
		h.storage.AddCounter(metrics.ID, *metrics.Delta)
	case models.Gauge:
		if metrics.Value == nil {
			return fmt.Errorf("no gauge value found for: %s", metrics.ID)
		}
		h.storage.SetGauge(metrics.ID, *metrics.Value)
	default:
		return fmt.Errorf("unknown metric type: %q", metrics.MType)
	}

	return nil
}

func isEmptyParam(p string) bool {
	return len(p) == 0
}

func newMetrics(id, mType, rawValue string) (*models.Metrics, error) {
	m := models.Metrics{
		ID:    id,
		MType: mType,
	}

	switch mType {
	case models.Counter:
		delta, err := strconv.ParseInt(rawValue, 10, 64)
		if err != nil {
			return nil, ErrParseValue
		}
		m.Delta = &delta
	case models.Gauge:
		value, err := strconv.ParseFloat(rawValue, 64)
		if err != nil {
			return nil, ErrParseValue
		}
		m.Value = &value
	default:
		return nil, ErrUnknownType
	}

	return &m, nil
}
