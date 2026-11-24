package services

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	models "github.com/SysoevAndrey/metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type runtimeStorageMock struct {
	counters map[string]int64
	gauges   map[string]float64
}

func newRuntimeStorageMock() *runtimeStorageMock {
	return &runtimeStorageMock{
		counters: map[string]int64{},
		gauges:   map[string]float64{},
	}
}

func (s *runtimeStorageMock) SetGauge(ctx context.Context, name string, value float64) error {
	s.gauges[name] = value
	return nil
}

func (s *runtimeStorageMock) AddCounter(ctx context.Context, name string, value int64) error {
	s.counters[name] += value
	return nil
}

func (s *runtimeStorageMock) GetAll(ctx context.Context) (models.Snapshot, error) {
	return models.Snapshot{
		Gauges:   s.gauges,
		Counters: s.counters,
	}, nil
}

type persistentStorageMock struct {
	buf *bytes.Buffer
}

func newPersistentStorageMock() *persistentStorageMock {
	return &persistentStorageMock{
		buf: bytes.NewBuffer(nil),
	}
}

func (s *persistentStorageMock) WriteMetrics(metrics []models.Metrics) error {
	return json.NewEncoder(s.buf).Encode(metrics)
}

func (s *persistentStorageMock) ReadMetrics() ([]models.Metrics, error) {
	metrics := []models.Metrics{}
	if err := json.NewDecoder(s.buf).Decode(&metrics); err != nil {
		return []models.Metrics{}, err
	}
	return metrics, nil
}

func TestBackup_SeedRuntimeStorageIfNeeded(t *testing.T) {
	t.Run("doesn't seed if Restore is false", func(t *testing.T) {
		rs := newRuntimeStorageMock()
		ps := newPersistentStorageMock()
		cfg := BackupConfig{Restore: false}

		ps.WriteMetrics([]models.Metrics{
			{ID: "PollCount", MType: models.Counter, Delta: new(int64(42))},
			{ID: "Alloc", MType: models.Gauge, Value: new(17.03)},
		})

		b := NewBackup(rs, ps, cfg)
		err := b.SeedRuntimeStorageIfNeeded(t.Context())
		require.NoError(t, err)

		metrics, err := b.rs.GetAll(t.Context())
		require.NoError(t, err)

		assert.Empty(t, metrics.Counters)
		assert.Empty(t, metrics.Gauges)
	})

	t.Run("seeds if Restore is true", func(t *testing.T) {
		rs := newRuntimeStorageMock()
		ps := newPersistentStorageMock()
		cfg := BackupConfig{Restore: true}

		err := ps.WriteMetrics([]models.Metrics{
			{ID: "PollCount", MType: models.Counter, Delta: new(int64(42))},
		})
		require.NoError(t, err)

		b := NewBackup(rs, ps, cfg)
		err = b.SeedRuntimeStorageIfNeeded(t.Context())
		require.NoError(t, err)

		metrics, err := b.rs.GetAll(t.Context())
		require.NoError(t, err)

		pollCount, found := metrics.Counters["PollCount"]
		assert.True(t, found)
		assert.Equal(t, int64(42), pollCount)
	})

	tests := []struct {
		name   string
		metric models.Metrics
	}{
		{
			name:   "throws if counter has no delta",
			metric: models.Metrics{ID: "PollCount", MType: models.Counter},
		},
		{
			name:   "throws if gauge has no value",
			metric: models.Metrics{ID: "Alloc", MType: models.Gauge},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := newRuntimeStorageMock()
			ps := newPersistentStorageMock()
			cfg := BackupConfig{Restore: true}

			err := ps.WriteMetrics([]models.Metrics{tt.metric})
			require.NoError(t, err)

			b := NewBackup(rs, ps, cfg)
			err = b.SeedRuntimeStorageIfNeeded(t.Context())
			assert.Error(t, err)
		})
	}
}

func TestBackup_createSnapshot(t *testing.T) {
	t.Run("creates snapshot", func(t *testing.T) {
		rs := newRuntimeStorageMock()
		ps := newPersistentStorageMock()
		cfg := BackupConfig{Restore: true}

		rs.AddCounter(t.Context(), "PollCount", int64(42))
		rs.SetGauge(t.Context(), "Alloc", 17.03)

		b := NewBackup(rs, ps, cfg)
		err := b.createSnapshot(t.Context())
		require.NoError(t, err)

		result, err := b.ps.ReadMetrics()
		require.NoError(t, err)

		metrics := []models.Metrics{
			{ID: "PollCount", MType: models.Counter, Delta: new(int64(42))},
			{ID: "Alloc", MType: models.Gauge, Value: new(17.03)},
		}

		assert.ElementsMatch(t, metrics, result)
	})
}
