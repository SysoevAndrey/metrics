package handlers

import (
	"context"

	models "github.com/SysoevAndrey/metrics/internal/model"
)

type storage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func newStorage() *storage {
	return &storage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (s *storage) AddCounter(ctx context.Context, name string, value int64) error {
	s.counters[name] += value
	return nil
}

func (s *storage) SetGauge(ctx context.Context, name string, value float64) error {
	s.gauges[name] = value
	return nil
}

func (s *storage) GetCounter(ctx context.Context, name string) (int64, error) {
	counter, exists := s.counters[name]
	if !exists {
		return 0, models.ErrNotFound
	}
	return counter, nil
}

func (s *storage) GetGauge(ctx context.Context, name string) (float64, error) {
	gauge, exists := s.gauges[name]
	if !exists {
		return 0, models.ErrNotFound
	}
	return gauge, nil
}

func (s *storage) GetAll(ctx context.Context) (models.Snapshot, error) {
	return models.Snapshot{
		Gauges:   s.gauges,
		Counters: s.counters,
	}, nil
}
