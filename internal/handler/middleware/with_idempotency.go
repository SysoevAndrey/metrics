package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"
)

const (
	requestTTL    = time.Minute
	clearInterval = 2 * time.Minute
)

type requestsStorage struct {
	requests map[string]requestData
	mu       sync.Mutex
}

func newRequestStorage() *requestsStorage {
	return &requestsStorage{
		requests: map[string]requestData{},
	}
}

type requestData struct {
	state     requestState
	code      int
	expiresAt time.Time
}

type requestState int

const (
	pending requestState = iota + 1
	processed
)

type idempotencyResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *idempotencyResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func WithIdempotency(ctx context.Context) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		storage := newRequestStorage()
		go clearStaleKeys(ctx, storage)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("Idempotency-Key")
			if len(key) == 0 {
				h.ServeHTTP(w, r)
				return
			}

			storage.mu.Lock()
			data, exists := storage.requests[key]
			if exists {
				switch data.state {
				case pending:
					storage.mu.Unlock()
					w.WriteHeader(http.StatusConflict)
					return
				case processed:
					storage.mu.Unlock()
					w.WriteHeader(data.code)
					return
				}
			}

			storage.requests[key] = requestData{state: pending}
			storage.mu.Unlock()

			defer func() {
				storage.mu.Lock()
				defer storage.mu.Unlock()
				data, exists := storage.requests[key]
				if !exists {
					return
				}
				if data.state == pending {
					delete(storage.requests, key)
				}
			}()

			writer := idempotencyResponseWriter{ResponseWriter: w}
			h.ServeHTTP(&writer, r)

			if writer.statusCode < 300 {
				storage.mu.Lock()
				storage.requests[key] = requestData{
					state:     processed,
					code:      writer.statusCode,
					expiresAt: time.Now().Add(requestTTL),
				}
				storage.mu.Unlock()
			}
		})
	}
}

func clearStaleKeys(ctx context.Context, storage *requestsStorage) {
	ticker := time.NewTicker(clearInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			storage.mu.Lock()
			for key, data := range storage.requests {
				if data.state == processed && t.After(data.expiresAt) {
					delete(storage.requests, key)
				}
			}
			storage.mu.Unlock()
		}
	}
}
