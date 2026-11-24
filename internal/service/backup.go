package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SysoevAndrey/metrics/internal/logger"
	models "github.com/SysoevAndrey/metrics/internal/model"
	"go.uber.org/zap"
)

type RuntimeStorage interface {
	SetGauge(ctx context.Context, name string, value float64) error
	AddCounter(ctx context.Context, name string, value int64) error

	GetAll(ctx context.Context) (models.Snapshot, error)
}

type PersistentStorage interface {
	WriteMetrics(metrics []models.Metrics) error
	ReadMetrics() ([]models.Metrics, error)
}

type Backup struct {
	rs  RuntimeStorage
	ps  PersistentStorage
	cfg BackupConfig

	mu sync.Mutex
}

type BackupConfig struct {
	Restore       bool
	StoreInterval time.Duration
}

func NewBackup(rs RuntimeStorage, ps PersistentStorage, cfg BackupConfig) *Backup {
	return &Backup{
		rs:  rs,
		ps:  ps,
		cfg: cfg,
	}
}

func (b *Backup) SeedRuntimeStorageIfNeeded(ctx context.Context) error {
	if !b.cfg.Restore {
		return nil
	}

	metrics, err := b.ps.ReadMetrics()
	if err != nil {
		return err
	}

	for _, m := range metrics {
		switch m.MType {
		case models.Counter:
			if m.Delta == nil {
				return fmt.Errorf("counter metric %s has no delta", m.ID)
			}
			if err = b.rs.AddCounter(ctx, m.ID, *m.Delta); err != nil {
				return err
			}
		case models.Gauge:
			if m.Value == nil {
				return fmt.Errorf("gauge metric %s has no value", m.ID)
			}
			if err = b.rs.SetGauge(ctx, m.ID, *m.Value); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown metric type: %s", m.MType)
		}
	}

	logger.Log.Info("seeded runtime storage")
	return nil
}

func (b *Backup) SetupPersistentStorageUpdateIfNeeded(ctx context.Context) {
	if b.cfg.StoreInterval <= 0 {
		return
	}

	for {
		time.Sleep(b.cfg.StoreInterval)
		if err := b.createSnapshot(ctx); err != nil {
			logger.Log.Error("failed to write metrics", zap.Error(err))
		}
	}
}

func (b *Backup) CreateSnapshotIfNeeded(ctx context.Context) error {
	if b.cfg.StoreInterval != 0 {
		return nil
	}
	return b.createSnapshot(ctx)
}

func (b *Backup) createSnapshot(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	sh, err := b.rs.GetAll(ctx)
	if err != nil {
		return err
	}

	metrics := []models.Metrics{}
	for name, value := range sh.Gauges {
		m := models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &value,
		}
		metrics = append(metrics, m)
	}
	for name, delta := range sh.Counters {
		m := models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &delta,
		}
		metrics = append(metrics, m)
	}

	if err := b.ps.WriteMetrics(metrics); err != nil {
		return err
	}

	logger.Log.Info("metrics were written")
	return nil
}
