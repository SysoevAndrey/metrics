package services

import (
	"bytes"
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

func (s *runtimeStorageMock) SetGauge(name string, value float64) {
	s.gauges[name] = value
}

func (s *runtimeStorageMock) AddCounter(name string, value int64) {
	s.counters[name] += value
}

func (s *runtimeStorageMock) GetAll() models.Snapshot {
	return models.Snapshot{
		Gauges:   s.gauges,
		Counters: s.counters,
	}
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
		err := b.SeedRuntimeStorageIfNeeded()
		require.NoError(t, err)

		assert.Empty(t, b.rs.GetAll().Counters)
		assert.Empty(t, b.rs.GetAll().Gauges)
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
		err = b.SeedRuntimeStorageIfNeeded()
		require.NoError(t, err)

		pollCount, found := b.rs.GetAll().Counters["PollCount"]
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
			err = b.SeedRuntimeStorageIfNeeded()
			assert.Error(t, err)
		})
	}
}

func TestBackup_createSnapshot(t *testing.T) {
	t.Run("creates snapshot", func(t *testing.T) {
		rs := newRuntimeStorageMock()
		ps := newPersistentStorageMock()
		cfg := BackupConfig{Restore: true}

		rs.AddCounter("PollCount", int64(42))
		rs.SetGauge("Alloc", 17.03)

		b := NewBackup(rs, ps, cfg)
		err := b.createSnapshot()
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
