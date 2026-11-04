package main

import (
	"time"

	"github.com/SysoevAndrey/metrics/internal/agent"
)

func main() {
	parseFlags()
	a := agent.NewAgent(
		time.Duration(config.PollInterval)*time.Second,
		time.Duration(config.ReportInterval)*time.Second,
		config.Host,
	)
	err := a.Run()
	if err != nil {
		panic(err)
	}
}
