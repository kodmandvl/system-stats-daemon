// Package diskload — tps и KB/s по дискам (iostat).
package diskload

import "github.com/kodmandvl/system-stats-daemon/internal/collector"

type Collector struct {
	exec collector.CommandExecutor
}

func New(exec collector.CommandExecutor) *Collector {
	return &Collector{exec: exec}
}

func (c *Collector) Name() string { return "disks_load" }
