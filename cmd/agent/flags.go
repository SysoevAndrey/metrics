package main

import (
	"flag"
	"time"
)

var config struct {
	host           string
	pollInterval   time.Duration
	reportInterval time.Duration
}

func parseFlags() {
	flag.StringVar(&config.host, "a", "localhost:8080", "target server host")
	flag.DurationVar(&config.pollInterval, "p", 2*time.Second, "poll interval")
	flag.DurationVar(&config.reportInterval, "r", 10*time.Second, "report interval")
	flag.Parse()
}
