// Package filesystem — df -k и df -i (общий код для Linux и Darwin).
package filesystem

import "github.com/kodmandvl/system-stats-daemon/internal/collector"

type Collector struct {
	exec collector.CommandExecutor
}

func New(exec collector.CommandExecutor) *Collector {
	return &Collector{exec: exec}
}

func (c *Collector) Name() string { return "filesystem" }
