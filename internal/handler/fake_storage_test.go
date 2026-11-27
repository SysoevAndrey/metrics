package handlers

import (
	"context"

	models "github.com/SysoevAndrey/metrics/internal/model"
)

type storage struct {
	gauges   map[string]float64
	counters map[string]int64
	err      error
}

func newStorage(err error) *storage {
	return &storage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
		err:      err,
	}
}

func (s *storage) AddCounter(ctx context.Context, name string, value int64) error {
	if s.err != nil {
		return s.err
	}
	s.counters[name] += value
	return nil
}

func (s *storage) SetGauge(ctx context.Context, name string, value float64) error {
	if s.err != nil {
		return s.err
	}
	s.gauges[name] = value
	return nil
}

func (s *storage) ApplyBatch(ctx context.Context, metrics []models.Metrics) error {
	if s.err != nil {
		return s.err
	}

	for _, metric := range metrics {
		switch metric.MType {
		case models.Counter:
			if metric.Delta == nil {
				return models.ErrMissingValue
			}
			if err := s.AddCounter(ctx, metric.ID, *metric.Delta); err != nil {
				return err
			}
		case models.Gauge:
			if metric.Value == nil {
				return models.ErrMissingValue
			}
			if err := s.SetGauge(ctx, metric.ID, *metric.Value); err != nil {
				return err
			}
		default:
			return models.ErrUnknownType
		}
	}

	return nil
}

func (s *storage) GetCounter(ctx context.Context, name string) (int64, error) {
	if s.err != nil {
		return 0, s.err
	}
	counter, exists := s.counters[name]
	if !exists {
		return 0, models.ErrNotFound
	}
	return counter, nil
}

func (s *storage) GetGauge(ctx context.Context, name string) (float64, error) {
	if s.err != nil {
		return 0, s.err
	}
	gauge, exists := s.gauges[name]
	if !exists {
		return 0, models.ErrNotFound
	}
	return gauge, nil
}

func (s *storage) GetAll(ctx context.Context) (models.Snapshot, error) {
	if s.err != nil {
		return models.Snapshot{}, s.err
	}
	return models.Snapshot{
		Gauges:   s.gauges,
		Counters: s.counters,
	}, nil
}
