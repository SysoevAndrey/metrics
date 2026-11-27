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
	timeout:      200 * time.Millisecond,
	retryCount:   3,
	retryMaxWait: time.Second,
	retryStep:    10 * time.Millisecond,
}

func TestAgent_sendMetrics(t *testing.T) {
	tests := []struct {
		name        string
		gauges      map[string]float64
		counters    map[string]int64
		mode        requestMode
		pathToThrow string
		thrownErr   int
		wantMode    requestMode
		wantCalls   int
	}{
		{
			name:      "sends all metrics sequentially",
			gauges:    map[string]float64{"Alloc": 17.03},
			counters:  map[string]int64{"PollCount": 42},
			mode:      sequential,
			wantMode:  sequential,
			wantCalls: 2,
		},
		{
			name:        "error doesn't block sending sequentially",
			gauges:      map[string]float64{"Alloc": 17.03},
			counters:    map[string]int64{"PollCount": 42},
			pathToThrow: "/update/counter/PollCount/42",
			thrownErr:   http.StatusBadRequest,
			mode:        sequential,
			wantMode:    sequential,
			wantCalls:   2,
		},
		{
			name:      "sends all metrics using batch",
			gauges:    map[string]float64{"Alloc": 17.03},
			counters:  map[string]int64{"PollCount": 42},
			mode:      batch,
			wantMode:  batch,
			wantCalls: 1,
		},
		{
			name:        "switches mode to sequential on 404 error",
			gauges:      map[string]float64{"Alloc": 17.03},
			counters:    map[string]int64{"PollCount": 42},
			pathToThrow: "/updates",
			thrownErr:   http.StatusNotFound,
			mode:        batch,
			wantMode:    sequential,
			wantCalls:   3,
		},
		{
			name:        "doesn't switch mode on 500 error",
			gauges:      map[string]float64{"Alloc": 17.03},
			counters:    map[string]int64{"PollCount": 42},
			pathToThrow: "/updates",
			thrownErr:   http.StatusInternalServerError,
			mode:        batch,
			wantMode:    batch,
			wantCalls:   4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls int

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Path == tt.pathToThrow {
					w.WriteHeader(tt.thrownErr)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			host, found := strings.CutPrefix(ts.URL, "http://")
			require.True(t, found)

			agent := NewAgent(0, 0, host)
			agent.setupClient(testClientConfig)

			agent.mode = tt.mode
			agent.sendMetrics(tt.counters, tt.gauges)

			assert.Equal(t, tt.wantMode, agent.mode)
			assert.Equal(t, tt.wantCalls, calls)
		})
	}
}

func TestAgent_sendRequest(t *testing.T) {
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
			keys := []string{}

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				keys = append(keys, r.Header.Get("Idempotency-Key"))
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
			err := agent.sendRequest("/", []byte{})

			assert.Equal(t, tt.wantCalls, calls)
			assert.Len(t, keys, tt.wantCalls)

			key := keys[0]
			assert.NotEmpty(t, key)
			for i := 1; i < len(keys); i++ {
				assert.Equal(t, key, keys[i])
			}

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	t.Run("changes key between requests", func(t *testing.T) {
		keys := []string{}

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			keys = append(keys, r.Header.Get("Idempotency-Key"))
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		host, found := strings.CutPrefix(ts.URL, "http://")
		require.True(t, found)

		agent := NewAgent(0, 0, host)
		agent.setupClient(testClientConfig)

		err := agent.sendRequest("/", []byte{})
		require.NoError(t, err)
		err = agent.sendRequest("/", []byte{})
		require.NoError(t, err)

		assert.Len(t, keys, 2)
		assert.NotEqual(t, keys[0], keys[1])
	})
}
