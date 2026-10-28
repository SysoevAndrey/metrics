package agent

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

type Agent struct {
	pollInterval   time.Duration
	reportInterval time.Duration
	host           string
	client         *resty.Client
}

type clientConfig struct {
	timeout          time.Duration
	retryCount       int
	retryWaitTime    time.Duration
	retryMaxWaitTime time.Duration
}

func NewAgent(pollInterval, reportInterval time.Duration, host string) *Agent {
	return &Agent{
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		host:           host,
		client:         resty.New(),
	}
}

func (a *Agent) Run() error {
	var wg sync.WaitGroup
	collector := newCollector()

	a.setupClient(clientConfig{
		timeout:          200 * time.Millisecond,
		retryCount:       3,
		retryWaitTime:    200 * time.Millisecond,
		retryMaxWaitTime: time.Second,
	})

	wg.Go(func() {
		for {
			collector.Collect()
			time.Sleep(a.pollInterval)
		}
	})

	wg.Go(func() {
		for {
			counters, gauges := collector.GetSnapshot()
			a.sendMetrics(counters, gauges)
			time.Sleep(a.reportInterval)
		}
	})

	wg.Wait()

	return nil
}

func (a *Agent) setupClient(config clientConfig) {
	a.client.
		SetBaseURL("http://" + a.host).
		SetTimeout(config.timeout).
		SetRetryCount(config.retryCount).
		SetRetryWaitTime(config.retryWaitTime).
		SetRetryMaxWaitTime(config.retryMaxWaitTime).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return err != nil || r.StatusCode() >= http.StatusInternalServerError
		})
}

type metric struct {
	t string
	n string
	v string
}

func (a *Agent) sendMetrics(counters map[string]int64, gauges map[string]float64) {
	metrics := make([]metric, 0, len(counters)+len(gauges))

	for name, value := range counters {
		metrics = append(metrics, metric{t: "counter", n: name, v: strconv.FormatInt(value, 10)})
	}

	for name, value := range gauges {
		metrics = append(metrics, metric{t: "gauge", n: name, v: strconv.FormatFloat(value, 'f', -1, 64)})
	}

	for _, m := range metrics {
		err := a.sendMetric(m)
		if err != nil {
			log.Print(err)
		}
	}
}

func (a *Agent) sendMetric(m metric) error {
	resp, err := a.client.R().
		SetPathParams(map[string]string{"t": m.t, "n": m.n, "v": m.v}).
		SetHeader("Content-Type", "text/plain").
		Post("/update/{t}/{n}/{v}")

	if err != nil {
		return fmt.Errorf("failed to send %s=%s: %v", m.n, m.v, err)
	}
	if resp.IsError() {
		return fmt.Errorf("failed to send %s=%s: %v", m.n, m.v, resp.Status())
	}
	return nil
}
