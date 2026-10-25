package handlers

import models "github.com/SysoevAndrey/metrics/internal/model"

type storage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func (s storage) AddCounter(name string, value int64) {}
func (s storage) SetGauge(name string, value float64) {}

func (s storage) GetCounter(name string) (int64, bool) {
	counter, exists := s.counters[name]
	return counter, exists
}

func (s storage) GetGauge(name string) (float64, bool) {
	gauge, exists := s.gauges[name]
	return gauge, exists
}

func (s storage) GetAll() models.Snapshot {
	return models.Snapshot{
		Gauges:   s.gauges,
		Counters: s.counters,
	}
}
