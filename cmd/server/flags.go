package main

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

var config struct {
	Host            string `env:"ADDRESS"`
	LogLevel        string `env:"LOG_LEVEL"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

func parseFlags() {
	flag.StringVar(&config.Host, "a", "localhost:8080", "host to run server")
	flag.StringVar(&config.LogLevel, "l", "info", "log level")
	flag.IntVar(&config.StoreInterval, "i", 300, "save to store interval")
	flag.StringVar(&config.FileStoragePath, "f", "./metrics_snapshot.json", "path to metrics storage file")
	flag.BoolVar(&config.Restore, "r", false, "restore data on startup")
	flag.Parse()

	err := env.Parse(&config)
	if err != nil {
		fmt.Println(err)
	}
}
