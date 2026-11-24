package handlers

import (
	"context"

	models "github.com/SysoevAndrey/metrics/internal/model"
)

type MetricStorage interface {
	SetGauge(ctx context.Context, name string, value float64) error
	AddCounter(ctx context.Context, name string, value int64) error

	ApplyBatch(ctx context.Context, metrics []models.Metrics) error

	GetGauge(ctx context.Context, name string) (float64, error)
	GetCounter(ctx context.Context, name string) (int64, error)

	GetAll(ctx context.Context) (models.Snapshot, error)
}
