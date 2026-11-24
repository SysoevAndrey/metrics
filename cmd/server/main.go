package main

import (
	"context"
	"log"
	"net/http"

	handlers "github.com/SysoevAndrey/metrics/internal/handler"
	"github.com/SysoevAndrey/metrics/internal/handler/middleware"
	"github.com/SysoevAndrey/metrics/internal/logger"
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

	ctx := context.Background()

	store, err := newStore()
	if err != nil {
		return err
	}
	if store.closer != nil {
		defer store.closer.Close()
	}

	hook, err := setupBackup(ctx, store.metrics)
	if err != nil {
		return err
	}

	router, err := newRouter(*store, hook)
	if err != nil {
		return err
	}

	logger.Log.Info("Server started", zap.String("address", config.Host))
	return http.ListenAndServe(config.Host, router)
}

func newRouter(store store, hook func(context.Context) error) (http.Handler, error) {
	r := chi.NewRouter()

	valueHandler := handlers.NewValueHandler(store.metrics)
	updateHandler := handlers.NewUpdateHandler(store.metrics)
	indexHandler, err := handlers.NewIndexHandler(store.metrics)
	if err != nil {
		return nil, err
	}

	r.Use(middleware.WithLogging)
	r.Use(middleware.WithCompression)

	r.Get("/", indexHandler.GetAll)

	r.Route("/update", func(r chi.Router) {
		if hook != nil {
			r.Use(middleware.WithPostHook(hook))
		}

		r.Post("/{type}/{name}/{value}", updateHandler.ByPath)
		r.Post("/", updateHandler.ByBody)
	})

	r.Route("/updates", func(r chi.Router) {
		if hook != nil {
			r.Use(middleware.WithPostHook(hook))
		}

		r.Post("/", updateHandler.Batch)
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", valueHandler.ByPath)
		r.Post("/", valueHandler.ByBody)
	})

	if store.pinger != nil {
		pingHandler := handlers.NewPingHandler(store.pinger)
		r.Get("/ping", pingHandler.Ping)
	}

	return r, nil
}
