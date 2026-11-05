package main

import (
	"net/http"

	handlers "github.com/SysoevAndrey/metrics/internal/handler"
	"github.com/SysoevAndrey/metrics/internal/handler/middleware"
	"github.com/SysoevAndrey/metrics/internal/logger"
	repositories "github.com/SysoevAndrey/metrics/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	parseFlags()

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	if err := logger.Initialize(config.LogLevel); err != nil {
		return err
	}

	ms := repositories.NewMemStorage()

	indexHandler, err := handlers.NewIndexHandler(ms)
	if err != nil {
		return err
	}

	valueHandler := handlers.NewValueHandler(ms)
	updateHandler := handlers.NewUpdateHandler(ms)

	r := chi.NewRouter()
	r.Use(middleware.WithLogging)
	r.Get("/", indexHandler.ServeHTTP)
	r.Get("/value/{type}/{name}", valueHandler.ServeHTTP)
	r.Post("/update/{type}/{name}/{value}", updateHandler.ServeHTTP)

	logger.Log.Info("Server started", zap.String("address", config.Host))
	return http.ListenAndServe(config.Host, r)
}
