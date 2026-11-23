package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/SysoevAndrey/metrics/internal/logger"
	models "github.com/SysoevAndrey/metrics/internal/model"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type ValueHandler struct {
	storage MetricStorage
}

func NewValueHandler(storage MetricStorage) *ValueHandler {
	return &ValueHandler{storage: storage}
}

func (h *ValueHandler) ByPath(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	requestMetrics := models.Metrics{
		ID:    metricName,
		MType: metricType,
	}

	metrics, err := h.value(r.Context(), requestMetrics)
	if errors.Is(err, models.ErrNotFound) {
		logger.Log.Debug(err.Error(), zap.String("id", requestMetrics.ID))
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if errors.Is(err, models.ErrUnknownType) {
		logger.Log.Debug(err.Error(), zap.String("id", requestMetrics.ID))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err != nil {
		logger.Log.Debug(err.Error(), zap.String("id", requestMetrics.ID))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	switch metrics.MType {
	case models.Counter:
		w.Write([]byte(strconv.FormatInt(*metrics.Delta, 10)))
	case models.Gauge:
		w.Write([]byte(strconv.FormatFloat(*metrics.Value, 'f', -1, 64)))
	}
	logger.Log.Debug("sending HTTP 200 response")
}

func (h *ValueHandler) ByBody(w http.ResponseWriter, r *http.Request) {
	var requestMetrics models.Metrics

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&requestMetrics); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metrics, err := h.value(r.Context(), requestMetrics)
	if errors.Is(err, models.ErrNotFound) {
		logger.Log.Debug(err.Error(), zap.String("id", requestMetrics.ID))
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if errors.Is(err, models.ErrUnknownType) {
		logger.Log.Debug(err.Error(), zap.String("id", requestMetrics.ID))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err != nil {
		logger.Log.Debug(err.Error(), zap.String("id", requestMetrics.ID))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	if err := enc.Encode(metrics); err != nil {
		logger.Log.Debug("error encoding response", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	logger.Log.Debug("sending HTTP 200 response")
}

func (h *ValueHandler) value(ctx context.Context, requestMetrics models.Metrics) (*models.Metrics, error) {
	switch requestMetrics.MType {
	case models.Counter:
		value, err := h.storage.GetCounter(ctx, requestMetrics.ID)
		if err != nil {
			return nil, err
		}
		metrics := models.Metrics{
			ID:    requestMetrics.ID,
			MType: requestMetrics.MType,
			Delta: &value,
		}
		return &metrics, nil
	case models.Gauge:
		value, err := h.storage.GetGauge(ctx, requestMetrics.ID)
		if err != nil {
			return nil, err
		}
		metrics := models.Metrics{
			ID:    requestMetrics.ID,
			MType: requestMetrics.MType,
			Value: &value,
		}
		return &metrics, nil
	default:
		return nil, models.ErrUnknownType
	}
}
