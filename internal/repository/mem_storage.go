package repositories

import (
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

func (ms *MemStorage) SetGauge(name string, value float64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.gauges[name] = value
}

func (ms *MemStorage) AddCounter(name string, value int64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.counters[name] += value
}

func (ms *MemStorage) GetCounter(name string) (int64, bool) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	counter, exists := ms.counters[name]
	return counter, exists
}

func (ms *MemStorage) GetGauge(name string) (float64, bool) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	gauge, exists := ms.gauges[name]
	return gauge, exists
}

func (ms *MemStorage) GetAll() models.Snapshot {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	sh := newSnapshot()

	maps.Copy(sh.Gauges, ms.gauges)
	maps.Copy(sh.Counters, ms.counters)

	return *sh
}
