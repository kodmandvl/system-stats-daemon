//go:build linux

// iostat -d -k на Linux.
package diskload

import (
	"bufio"
	"context"
	"strconv"
	"strings"

	"github.com/kodmandvl/system-stats-daemon/internal/model"
)

// Collect парсит таблицу iostat (строки устройств после заголовка).
func (c *Collector) Collect(ctx context.Context) (any, error) {
	out, err := c.exec.Run(ctx, "iostat", "-d", "-k")
	if err != nil {
		return nil, err
	}

	disks := make(map[string]*model.DiskLoad)
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		if lineNo <= 3 {
			continue
		}
		fields := strings.Fields(scanner.Text())
		if len(fields) < 4 {
			continue
		}
		tps, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			return nil, err
		}
		readKB, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			return nil, err
		}
		writeKB, err := strconv.ParseFloat(fields[3], 64)
		if err != nil {
			return nil, err
		}
		disks[fields[0]] = &model.DiskLoad{Tps: tps, Kbs: readKB + writeKB}
	}

	return &model.DisksLoadStats{Disks: disks}, nil
}
