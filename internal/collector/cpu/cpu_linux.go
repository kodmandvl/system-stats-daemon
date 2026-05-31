//go:build linux

// Реализация cpu для Linux (top batch mode).
package cpu

import (
	"context"
	"regexp"
	"strconv"

	"github.com/kodmandvl/system-stats-daemon/internal/model"
)

// Пример: %Cpu(s):  1.2 us,  0.5 sy,  0.0 ni, 98.3 id, ...
var cpuLine = regexp.MustCompile(`%Cpu\(s\):\s*([\d.]+)\s*us,\s*([\d.]+)\s*sy,\s*[\d.]+\s*ni,\s*([\d.]+)\s*id`)

// Collect запускает top в batch-режиме и извлекает us/sy/id.
func (c *Collector) Collect(ctx context.Context) (any, error) {
	out, err := c.exec.Run(ctx, "top", "-b", "-n", "1")
	if err != nil {
		return nil, err
	}

	m := cpuLine.FindSubmatch(out)
	if len(m) != 4 {
		return nil, model.ErrStatsNotValid
	}

	user, err := strconv.ParseFloat(string(m[1]), 64)
	if err != nil {
		return nil, err
	}
	system, err := strconv.ParseFloat(string(m[2]), 64)
	if err != nil {
		return nil, err
	}
	idle, err := strconv.ParseFloat(string(m[3]), 64)
	if err != nil {
		return nil, err
	}

	return &model.CPUStats{UserMode: user, SystemMode: system, Idle: idle}, nil
}
