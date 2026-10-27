package handlers

import models "github.com/SysoevAndrey/metrics/internal/model"

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

func (s *storage) AddCounter(name string, value int64) {
	s.counters[name] += value
}

func (s *storage) SetGauge(name string, value float64) {
	s.gauges[name] = value
}

func (s *storage) GetCounter(name string) (int64, bool) {
	counter, exists := s.counters[name]
	return counter, exists
}

func (s *storage) GetGauge(name string) (float64, bool) {
	gauge, exists := s.gauges[name]
	return gauge, exists
}

func (s *storage) GetAll() models.Snapshot {
	return models.Snapshot{
		Gauges:   s.gauges,
		Counters: s.counters,
	}
}
