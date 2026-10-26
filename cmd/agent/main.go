package main

import (
	"github.com/SysoevAndrey/metrics/internal/agent"
)

func main() {
	parseFlags()
	a := agent.NewAgent(config.pollInterval, config.reportInterval, config.host)
	err := a.Run()
	if err != nil {
		panic(err)
	}
}
