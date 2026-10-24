package agent

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_collector_Collect(t *testing.T) {
	callCount := 4
	metrics := map[string]struct{}{
		"Alloc":         {},
		"BuckHashSys":   {},
		"Frees":         {},
		"GCCPUFraction": {},
		"GCSys":         {},
		"HeapAlloc":     {},
		"HeapIdle":      {},
		"HeapInuse":     {},
		"HeapObjects":   {},
		"HeapReleased":  {},
		"HeapSys":       {},
		"LastGC":        {},
		"Lookups":       {},
		"MCacheInuse":   {},
		"MCacheSys":     {},
		"MSpanInuse":    {},
		"MSpanSys":      {},
		"Mallocs":       {},
		"NextGC":        {},
		"NumForcedGC":   {},
		"NumGC":         {},
		"OtherSys":      {},
		"PauseTotalNs":  {},
		"StackInuse":    {},
		"StackSys":      {},
		"Sys":           {},
		"TotalAlloc":    {},
		"RandomValue":   {},
	}

	c := newCollector()

	for range callCount {
		c.Collect()
	}

	assert.Equal(t, callCount, int(c.counters["PollCount"]))
	assert.Equal(t, len(metrics), len(c.gauges))

	for metric := range metrics {
		_, ok := c.gauges[metric]
		assert.True(t, ok, "Missing metric: %s", metric)
	}
}

func Test_collector_GetSnapshot(t *testing.T) {
	c := newCollector()

	c.Collect()
	counters, gauges := c.GetSnapshot()

	assert.EqualValues(t, c.counters, counters)
	assert.EqualValues(t, c.gauges, gauges)

	assert.NotEqual(t, reflect.ValueOf(c.counters).Pointer(), reflect.ValueOf(counters).Pointer())
	assert.NotEqual(t, reflect.ValueOf(c.gauges).Pointer(), reflect.ValueOf(gauges).Pointer())

	c.Collect()

	assert.NotEqualValues(t, c.counters, counters)
	assert.NotEqualValues(t, c.gauges, gauges)
}
