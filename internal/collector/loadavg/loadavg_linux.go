//go:build linux

package loadavg

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/kodmandvl/system-stats-daemon/internal/model"
)

// Collect читает /proc/loadavg — быстрее, чем парсить top.
func (c *Collector) Collect(ctx context.Context) (any, error) {
	_ = ctx
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return nil, err
	}

	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return nil, model.ErrStatsNotValid
	}

	la1, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return nil, err
	}
	la5, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return nil, err
	}
	la15, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return nil, err
	}

	return &model.LoadAvgStats{
		LoadAvg1:  la1,
		LoadAvg5:  la5,
		LoadAvg15: la15,
	}, nil
}
