package main

import (
	"github.com/SysoevAndrey/metrics/internal/agent"
	config "github.com/SysoevAndrey/metrics/internal/config/agent"
)

func main() {
	a := agent.NewAgent(config.PollInterval, config.ReportInterval, config.Host)
	err := a.Run()
	if err != nil {
		panic(err)
	}
}
