package repositories

import (
	"reflect"
	"testing"

	models "github.com/SysoevAndrey/metrics/internal/model"
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

func TestMemStorage_GetAll(t *testing.T) {
	tests := []struct {
		name      string
		initState state
		want      models.Snapshot
	}{
		{
			name:      "positive",
			initState: state{gauges: map[string]float64{"Alloc": 42.3}, counters: map[string]int64{"PollCount": 17}},
			want:      models.Snapshot{Gauges: map[string]float64{"Alloc": 42.3}, Counters: map[string]int64{"PollCount": 17}},
		},
		{
			name:      "keep separate values",
			initState: state{gauges: map[string]float64{"PollCount": 42}, counters: map[string]int64{"PollCount": 17}},
			want:      models.Snapshot{Gauges: map[string]float64{"PollCount": 42}, Counters: map[string]int64{"PollCount": 17}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := MemStorage{
				gauges:   tt.initState.gauges,
				counters: tt.initState.counters,
			}
			sh := ms.GetAll()

			t.Log("========", sh)

			assert.EqualValues(t, tt.want.Gauges, sh.Gauges)
			assert.EqualValues(t, tt.want.Counters, sh.Counters)

			assert.NotEqual(t, reflect.ValueOf(tt.want.Gauges).Pointer(), reflect.ValueOf(sh.Gauges).Pointer())
			assert.NotEqual(t, reflect.ValueOf(tt.want.Counters).Pointer(), reflect.ValueOf(sh.Counters).Pointer())
		})
	}
}
