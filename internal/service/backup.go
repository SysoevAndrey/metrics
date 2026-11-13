package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/SysoevAndrey/metrics/internal/logger"
	models "github.com/SysoevAndrey/metrics/internal/model"
	"go.uber.org/zap"
)

type RuntimeStorage interface {
	SetGauge(name string, value float64)
	AddCounter(name string, value int64)

	GetAll() models.Snapshot
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

func (b *Backup) SeedRuntimeStorageIfNeeded() error {
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
			b.rs.AddCounter(m.ID, *m.Delta)
		case models.Gauge:
			if m.Value == nil {
				return fmt.Errorf("gauge metric %s has no value", m.ID)
			}
			b.rs.SetGauge(m.ID, *m.Value)
		default:
			return fmt.Errorf("unknown metric type: %s", m.MType)
		}
	}

	logger.Log.Info("seeded runtime storage")
	return nil
}

func (b *Backup) SetupPersistentStorageUpdateIfNeeded() {
	if b.cfg.StoreInterval <= 0 {
		return
	}

	for {
		time.Sleep(b.cfg.StoreInterval)
		if err := b.createSnapshot(); err != nil {
			logger.Log.Error("failed to write metrics", zap.Error(err))
		}
	}
}

func (b *Backup) CreateSnapshotIfNeeded() error {
	if b.cfg.StoreInterval != 0 {
		return nil
	}
	return b.createSnapshot()
}

func (b *Backup) createSnapshot() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	sh := b.rs.GetAll()

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
