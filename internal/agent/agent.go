package agent

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Agent struct {
	pollInterval   time.Duration
	reportInterval time.Duration
	host           string
}

func NewAgent(pollInterval, reportInterval time.Duration, host string) *Agent {
	return &Agent{
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		host:           host,
	}
}

func (a *Agent) Run() error {
	var wg sync.WaitGroup
	collector := newCollector()
	client := &http.Client{
		Timeout: 200 * time.Millisecond,
	}

	wg.Go(func() {
		for {
			collector.Collect()
			time.Sleep(a.pollInterval)
		}
	})

	wg.Go(func() {
		for {
			counters, gauges := collector.GetSnapshot()
			a.sendMetrics(client, counters, gauges)
			time.Sleep(a.reportInterval)
		}
	})

	wg.Wait()

	return nil
}

func (a *Agent) sendMetrics(client *http.Client, counters map[string]int64, gauges map[string]float64) {
	type metric struct {
		t string
		n string
		v string
	}

	metrics := make([]metric, 0, len(counters)+len(gauges))

	for name, value := range counters {
		metrics = append(metrics, metric{t: "counter", n: name, v: strconv.FormatInt(value, 10)})
	}

	for name, value := range gauges {
		metrics = append(metrics, metric{t: "gauge", n: name, v: strconv.FormatFloat(value, 'f', -1, 64)})
	}

	for _, m := range metrics {
		request, err := a.newUpdateRequest(m.t, m.n, m.v)
		if err != nil {
			continue
		}
		statusCode, err := a.sendUpdateRequest(client, request)
		if err != nil || !(statusCode >= http.StatusOK && statusCode < http.StatusBadRequest) {
			return
		}
	}
}

func (a *Agent) newUpdateRequest(metricType, metricName, metricValue string) (*http.Request, error) {
	endpoint := fmt.Sprintf("%s/update/%s/%s/%s", a.host, metricType, metricName, metricValue)
	request, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "text/plain")
	return request, nil
}

func (a *Agent) sendUpdateRequest(client *http.Client, request *http.Request) (int, error) {
	response, err := client.Do(request)
	if err != nil {
		return 0, err
	}
	io.Copy(io.Discard, response.Body)
	response.Body.Close()
	return response.StatusCode, nil
}
