//go:build darwin

// Реализация cpu для macOS.
package cpu

import (
	"context"
	"regexp"
	"strconv"

	"github.com/kodmandvl/system-stats-daemon/internal/model"
)

// Пример macOS: CPU usage: 3.41% user, 6.82% sys, 89.77% idle.
var cpuLine = regexp.MustCompile(`CPU usage:\s*([\d.]+)% user,\s*([\d.]+)% sys,\s*([\d.]+)% idle`)

func (c *Collector) Collect(ctx context.Context) (any, error) {
	out, err := c.exec.Run(ctx, "top", "-l", "1", "-n", "0")
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
