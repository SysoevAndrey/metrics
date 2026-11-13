package main

import (
	"log"
	"net/http"
	"time"

	handlers "github.com/SysoevAndrey/metrics/internal/handler"
	"github.com/SysoevAndrey/metrics/internal/handler/middleware"
	"github.com/SysoevAndrey/metrics/internal/logger"
	repositories "github.com/SysoevAndrey/metrics/internal/repository"
	services "github.com/SysoevAndrey/metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	if err := parseFlags(); err != nil {
		log.Fatal(err)
	}

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	if err := logger.Initialize(config.LogLevel); err != nil {
		return err
	}

	ms := repositories.NewMemStorage()
	fs, err := repositories.NewFileStorage(config.FileStoragePath)
	if err != nil {
		return err
	}

	bc := services.BackupConfig{
		Restore:       config.Restore,
		StoreInterval: time.Duration(config.StoreInterval) * time.Second,
	}
	b := services.NewBackup(ms, fs, bc)
	if err := b.SeedRuntimeStorageIfNeeded(); err != nil {
		logger.Log.Error("failed to seed runtime storage", zap.Error(err))
	}
	go b.SetupPersistentStorageUpdateIfNeeded()

	indexHandler, err := handlers.NewIndexHandler(ms)
	if err != nil {
		return err
	}

	valueHandler := handlers.NewValueHandler(ms)
	updateHandler := handlers.NewUpdateHandler(ms)

	r := chi.NewRouter()
	r.Use(middleware.WithLogging)
	r.Use(middleware.WithCompression)
	r.Get("/", indexHandler.ServeHTTP)
	r.Route("/update", func(r chi.Router) {
		r.Use(middleware.WithPostHook(b.CreateSnapshotIfNeeded))
		r.Post("/{type}/{name}/{value}", updateHandler.ByPath)
		r.Post("/", updateHandler.ByBody)
	})
	r.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", valueHandler.ByPath)
		r.Post("/", valueHandler.ByBody)
	})

	logger.Log.Info("Server started", zap.String("address", config.Host))
	return http.ListenAndServe(config.Host, r)
}
