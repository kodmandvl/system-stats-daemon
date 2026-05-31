//go:build darwin

// iostat -d на macOS.
package diskload

import (
	"bufio"
	"context"
	"strconv"
	"strings"

	"github.com/kodmandvl/system-stats-daemon/internal/model"
)

// Collect читает упрощённый формат iostat BSD.
func (c *Collector) Collect(ctx context.Context) (any, error) {
	out, err := c.exec.Run(ctx, "iostat", "-d")
	if err != nil {
		return nil, err
	}

	disks := make(map[string]*model.DiskLoad)
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 || strings.HasPrefix(fields[0], "device") {
			continue
		}
		tps, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			continue
		}
		kbs, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			continue
		}
		disks[fields[0]] = &model.DiskLoad{Tps: tps, Kbs: kbs}
	}

	return &model.DisksLoadStats{Disks: disks}, nil
}
