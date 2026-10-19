package main

import (
	"net/http"

	handlers "github.com/SysoevAndrey/metrics/internal/handler"
	repositories "github.com/SysoevAndrey/metrics/internal/repository"
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

	mux := http.NewServeMux()
	mux.Handle("POST /update/{type}/{name}/{value}", updateHandler)

	return http.ListenAndServe(`:8080`, mux)
}
