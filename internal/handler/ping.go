package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/SysoevAndrey/metrics/internal/logger"
	"go.uber.org/zap"
)

type Pinger interface {
	PingContext(context.Context) error
}

type PingHandler struct {
	pinger Pinger
}

func NewPingHandler(p Pinger) *PingHandler {
	return &PingHandler{pinger: p}
}

func (ph *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err := ph.pinger.PingContext(ctx); err != nil {
		logger.Log.Error("failed to ping db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Log.Debug("successfully pinged db")
	w.WriteHeader(http.StatusOK)
}
