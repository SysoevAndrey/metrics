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

	updateHandler := handlers.UpdateHandler{
		Storage: ms,
	}

	r := chi.NewRouter()
	r.Use(handlers.WithPathValidation)
	r.Post("/update/{type}/{name}/{value}", updateHandler.ServeHTTP)

	return http.ListenAndServe(`:8080`, r)
}
