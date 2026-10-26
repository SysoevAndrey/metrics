package main

import (
	"flag"
)

var config struct {
	host           string
	pollInterval   int
	reportInterval int
}

func parseFlags() {
	flag.StringVar(&config.host, "a", "localhost:8080", "target server host")
	flag.IntVar(&config.pollInterval, "p", 2, "poll interval in seconds")
	flag.IntVar(&config.reportInterval, "r", 10, "report interval in seconds")
	flag.Parse()
}
