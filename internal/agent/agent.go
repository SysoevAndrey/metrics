package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	models "github.com/SysoevAndrey/metrics/internal/model"
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

func (a *Agent) sendMetrics(counters map[string]int64, gauges map[string]float64) {
	metrics := make([]models.Metrics, 0, len(counters)+len(gauges))

	for name, value := range counters {
		metrics = append(metrics, models.Metrics{ID: name, MType: models.Counter, Delta: &value})
	}

	for name, value := range gauges {
		metrics = append(metrics, models.Metrics{ID: name, MType: models.Gauge, Value: &value})
	}

	for _, m := range metrics {
		err := a.sendMetric(m)
		if err != nil {
			log.Print(err)
		}
	}
}

func (a *Agent) sendMetric(m models.Metrics) error {
	bodyRaw, err := json.Marshal(m)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	zw := gzip.NewWriter(buf)
	if _, err = zw.Write(bodyRaw); err != nil {
		return err
	}
	if err = zw.Close(); err != nil {
		return err
	}

	resp, err := a.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(buf.Bytes()).
		Post("/update")

	if err != nil {
		return fmt.Errorf("failed to send %s: %v", m.ID, err)
	}
	if resp.IsError() {
		return fmt.Errorf("failed to send %s: %v", m.ID, resp.Status())
	}
	return nil
}
