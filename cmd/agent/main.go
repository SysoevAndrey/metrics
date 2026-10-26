package main

import (
	"time"

	"github.com/SysoevAndrey/metrics/internal/agent"
)

func main() {
	parseFlags()
	a := agent.NewAgent(
		time.Duration(config.pollInterval)*time.Second,
		time.Duration(config.reportInterval)*time.Second,
		config.host,
	)
	err := a.Run()
	if err != nil {
		panic(err)
	}
}
