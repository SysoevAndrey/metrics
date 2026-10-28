package handlers

import models "github.com/SysoevAndrey/metrics/internal/model"

type MetricStorage interface {
	SetGauge(name string, value float64)
	AddCounter(name string, value int64)

	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)

	GetAll() models.Snapshot
}
