package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type state struct {
	gauges   map[string]float64
	counters map[string]int64
}

func TestMemStorage_SetGauge(t *testing.T) {
	tests := []struct {
		name      string
		initState state
		key       string
		value     float64
	}{
		{
			name:      "sets new value",
			initState: state{gauges: map[string]float64{}, counters: map[string]int64{}},
			key:       "Alloc",
			value:     42,
		},
		{
			name:      "replaces existing value",
			initState: state{gauges: map[string]float64{"Alloc": 33.3}, counters: map[string]int64{}},
			key:       "Alloc",
			value:     42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := MemStorage{
				gauges:   tt.initState.gauges,
				counters: tt.initState.counters,
			}
			ms.SetGauge(tt.key, tt.value)
			assert.Equal(t, tt.value, ms.gauges[tt.key])
		})
	}
}

func TestMemStorage_AddCounter(t *testing.T) {
	tests := []struct {
		name      string
		initState state
		key       string
		value     int64
		want      int64
	}{
		{
			name:      "adds new value",
			initState: state{gauges: map[string]float64{}, counters: map[string]int64{}},
			key:       "PollCount",
			value:     42,
			want:      42,
		},
		{
			name:      "updates existing value",
			initState: state{gauges: map[string]float64{}, counters: map[string]int64{"PollCount": 17}},
			key:       "PollCount",
			value:     42,
			want:      59,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := MemStorage{
				gauges:   tt.initState.gauges,
				counters: tt.initState.counters,
			}
			ms.AddCounter(tt.key, tt.value)
			assert.Equal(t, tt.want, ms.counters[tt.key])
		})
	}
}
