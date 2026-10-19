package repositories

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (ms *MemStorage) SetGauge(name string, value float64) {
	ms.gauges[name] = value
}

func (ms *MemStorage) AddCounter(name string, value int64) {
	ms.counters[name] += value
}

func (ms *MemStorage) GetCounter(name string) (int64, bool) {
	counter, exists := ms.counters[name]
	return counter, exists
}

func (ms *MemStorage) GetGauge(name string) (float64, bool) {
	gauge, exists := ms.gauges[name]
	return gauge, exists
}
