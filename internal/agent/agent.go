package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	models "github.com/SysoevAndrey/metrics/internal/model"
	"github.com/go-resty/resty/v2"
)

type requestMode int

const (
	batch requestMode = iota
	sequential
)

type Agent struct {
	pollInterval   time.Duration
	reportInterval time.Duration
	host           string
	client         *resty.Client
	mode           requestMode
}

type clientConfig struct {
	timeout      time.Duration
	retryCount   int
	retryMaxWait time.Duration
	retryStep    time.Duration
}

type APIError struct {
	StatusCode int
}

func (ae *APIError) Error() string {
	return fmt.Sprintf("status: %d", ae.StatusCode)
}

func NewAgent(pollInterval, reportInterval time.Duration, host string) *Agent {
	return &Agent{
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		host:           host,
		client:         resty.New(),
		mode:           batch,
	}
}

func (a *Agent) Run() error {
	var wg sync.WaitGroup
	collector := newCollector()

	a.setupClient(clientConfig{
		timeout:      200 * time.Millisecond,
		retryCount:   3,
		retryMaxWait: 5 * time.Second,
		retryStep:    time.Second,
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
		SetRetryMaxWaitTime(config.retryMaxWait).
		SetRetryAfter(func(c *resty.Client, r *resty.Response) (time.Duration, error) {
			return 2*time.Duration(r.Request.Attempt)*config.retryStep - config.retryStep, nil
		}).
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

	switch a.mode {
	case batch:
		a.sendMetricsBatch(metrics)
	case sequential:
		a.sendMetricsSequential(metrics)
	default:
		log.Printf("unknown mode: %d", a.mode)
	}
}

func (a *Agent) sendMetricsBatch(metrics []models.Metrics) {
	if len(metrics) == 0 {
		log.Println("skipping request: metrics empty")
		return
	}

	body, err := compressBody(metrics)
	if err != nil {
		log.Printf("failed to compress metrics: %v", err)
		return
	}

	err = a.sendRequest("/updates", body)

	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
		log.Println("got 404 error. switching to sequential mode")
		a.mode = sequential
		a.sendMetricsSequential(metrics)
	}
	if err != nil {
		log.Println(err)
	}
}

func (a *Agent) sendMetricsSequential(metrics []models.Metrics) {
	for _, metric := range metrics {
		body, err := compressBody(metric)
		if err != nil {
			log.Printf("failed to compress metrics: %v", err)
			continue
		}
		if err = a.sendRequest("/update", body); err != nil {
			log.Println(err)
		}
	}
}

func (a *Agent) sendRequest(path string, body []byte) error {
	resp, err := a.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Idempotency-Key", rand.Text()).
		SetBody(body).
		Post(path)

	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("request error: %w", &APIError{StatusCode: resp.StatusCode()})
	}
	return nil
}

func compressBody(body any) ([]byte, error) {
	bodyRaw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	zw := gzip.NewWriter(buf)
	if _, err = zw.Write(bodyRaw); err != nil {
		return nil, err
	}
	if err = zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
