package main

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

var config struct {
	Host           string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
}

func parseFlags() {
	flag.StringVar(&config.Host, "a", "localhost:8080", "target server host")
	flag.IntVar(&config.PollInterval, "p", 2, "poll interval in seconds")
	flag.IntVar(&config.ReportInterval, "r", 10, "report interval in seconds")
	flag.Parse()

	err := env.Parse(&config)
	if err != nil {
		fmt.Println(err)
	}
}
