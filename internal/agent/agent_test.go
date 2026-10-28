package agent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testClientConfig = clientConfig{
	timeout:          200 * time.Millisecond,
	retryCount:       3,
	retryWaitTime:    10 * time.Millisecond,
	retryMaxWaitTime: 50 * time.Millisecond,
}

func TestAgent_sendMetrics(t *testing.T) {
	tests := []struct {
		name        string
		gauges      map[string]float64
		counters    map[string]int64
		pathToThrow string
	}{
		{
			name:     "sends all metrics",
			gauges:   map[string]float64{"Alloc": 17.03},
			counters: map[string]int64{"PollCount": 42},
		},
		{
			name:        "error doesn't block sending",
			gauges:      map[string]float64{"Alloc": 17.03},
			counters:    map[string]int64{"PollCount": 42},
			pathToThrow: "/update/counter/PollCount/42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls int

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Path == tt.pathToThrow {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			host, found := strings.CutPrefix(ts.URL, "http://")
			require.True(t, found)

			agent := NewAgent(0, 0, host)
			agent.setupClient(testClientConfig)

			agent.sendMetrics(tt.counters, tt.gauges)
			assert.Equal(t, len(tt.counters)+len(tt.gauges), calls)
		})
	}
}

func TestAgent_sendMetric(t *testing.T) {
	tests := []struct {
		name       string
		failCount  int
		failStatus int
		wantErr    bool
		wantCalls  int
	}{
		{name: "ok without retry", failCount: 0, failStatus: 0, wantErr: false, wantCalls: 1},
		{name: "retry after 500, then ok", failCount: 1, failStatus: 500, wantErr: false, wantCalls: 2},
		{name: "throws error after retries", failCount: 4, failStatus: 500, wantErr: true, wantCalls: 4},
		{name: "4xx is not retried", failCount: 1, failStatus: 400, wantErr: true, wantCalls: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls int

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if calls <= tt.failCount {
					w.WriteHeader(tt.failStatus)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			host, found := strings.CutPrefix(ts.URL, "http://")
			require.True(t, found)

			agent := NewAgent(0, 0, host)
			agent.setupClient(testClientConfig)
			err := agent.sendMetric(metric{t: "counter", n: "PollCount", v: "42"})

			assert.Equal(t, tt.wantCalls, calls)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
