//go:build darwin

package loadavg

import (
	"context"
	"strconv"
	"strings"

	"github.com/kodmandvl/system-stats-daemon/internal/model"
)

// Collect использует uptime (кросс-платформенный вариант для macOS).
func (c *Collector) Collect(ctx context.Context) (any, error) {
	out, err := c.exec.Run(ctx, "uptime")
	if err != nil {
		return nil, err
	}

	fields := strings.Fields(string(out))
	if len(fields) < 3 {
		return nil, model.ErrStatsNotValid
	}

	parse := func(s string) (float64, error) {
		return strconv.ParseFloat(strings.ReplaceAll(s, ",", "."), 64)
	}

	la1, err := parse(fields[len(fields)-3])
	if err != nil {
		return nil, err
	}
	la5, err := parse(fields[len(fields)-2])
	if err != nil {
		return nil, err
	}
	la15, err := parse(fields[len(fields)-1])
	if err != nil {
		return nil, err
	}

	return &model.LoadAvgStats{LoadAvg1: la1, LoadAvg5: la5, LoadAvg15: la15}, nil
}
