// Package cpu собирает загрузку CPU через top (формат вывода зависит от ОС).
package cpu

import "github.com/kodmandvl/system-stats-daemon/internal/collector"

// Collector парсит вывод top.
type Collector struct {
	exec collector.CommandExecutor
}

// New создаёт коллектор CPU.
func New(exec collector.CommandExecutor) *Collector {
	return &Collector{exec: exec}
}

func (c *Collector) Name() string { return "cpu" }
