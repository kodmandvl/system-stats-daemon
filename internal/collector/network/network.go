// Package network — сетевая статистика (реализация: network_linux.go / network_darwin.go).
package network

import "github.com/kodmandvl/system-stats-daemon/internal/collector"

// Collector собирает сетевую статистику (top talkers, сокеты, TCP states).
type Collector struct {
	exec collector.CommandExecutor
}

// New создаёт сетевой коллектор.
func New(exec collector.CommandExecutor) *Collector {
	return &Collector{exec: exec}
}

func (c *Collector) Name() string { return "network" }
