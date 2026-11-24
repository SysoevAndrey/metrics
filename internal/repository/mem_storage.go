package repositories

import (
	"context"
	"maps"
	"sync"

	models "github.com/SysoevAndrey/metrics/internal/model"
)

func newSnapshot() *models.Snapshot {
	return &models.Snapshot{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
}

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64

	mu sync.Mutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (ms *MemStorage) SetGauge(ctx context.Context, name string, value float64) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.gauges[name] = value
	return nil
}

func (ms *MemStorage) AddCounter(ctx context.Context, name string, value int64) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.counters[name] += value
	return nil
}

func (s *MemStorage) ApplyBatch(ctx context.Context, metrics []models.Metrics) error {
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

func (ms *MemStorage) GetCounter(ctx context.Context, name string) (int64, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	counter, exists := ms.counters[name]
	if !exists {
		return 0, models.ErrNotFound
	}
	return counter, nil
}

func (ms *MemStorage) GetGauge(ctx context.Context, name string) (float64, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	gauge, exists := ms.gauges[name]
	if !exists {
		return 0, models.ErrNotFound
	}
	return gauge, nil
}

func (ms *MemStorage) GetAll(ctx context.Context) (models.Snapshot, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	sh := newSnapshot()

	maps.Copy(sh.Gauges, ms.gauges)
	maps.Copy(sh.Counters, ms.counters)

	return *sh, nil
}
