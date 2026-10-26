package main

import (
	"net/http"

	handlers "github.com/SysoevAndrey/metrics/internal/handler"
	repositories "github.com/SysoevAndrey/metrics/internal/repository"
	"github.com/go-chi/chi/v5"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	ms := repositories.NewMemStorage()

	indexHandler, err := handlers.NewIndexHandler(ms)
	if err != nil {
		return err
	}

	valueHandler := handlers.NewValueHandler(ms)
	updateHandler := handlers.NewUpdateHandler(ms)

	r := chi.NewRouter()
	r.Get("/", indexHandler.ServeHTTP)
	r.Get("/value/{type}/{name}", valueHandler.ServeHTTP)
	r.Post("/update/{type}/{name}/{value}", updateHandler.ServeHTTP)

	return http.ListenAndServe(`:8080`, r)
}
