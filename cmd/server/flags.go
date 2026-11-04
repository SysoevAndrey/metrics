package main

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

var config struct {
	Host string `env:"ADDRESS"`
}

func parseFlags() {
	flag.StringVar(&config.Host, "a", "localhost:8080", "host to run server")
	flag.Parse()

	err := env.Parse(&config)
	if err != nil {
		fmt.Println(err)
	}
}
