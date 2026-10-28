package main

import "flag"

var host string

func parseFlags() {
	flag.StringVar(&host, "a", "localhost:8080", "host to run server")
	flag.Parse()
}
