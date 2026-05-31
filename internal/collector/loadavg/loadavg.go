// Package loadavg собирает load average (linux: /proc, darwin: uptime).
package loadavg

import "github.com/kodmandvl/system-stats-daemon/internal/collector"

// Collector собирает load average.
type Collector struct {
	exec collector.CommandExecutor
}

// New создаёт коллектор с исполнителем команд (на Linux exec не используется).
func New(exec collector.CommandExecutor) *Collector {
	return &Collector{exec: exec}
}

// Name — идентификатор для логов и отладки.
func (c *Collector) Name() string { return "load_avg" }
