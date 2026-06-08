// Сбор и парсинг df — см. package filesystem (fs.go).
package filesystem

import (
	"bufio"
	"context"
	"strconv"
	"strings"

	"github.com/kodmandvl/system-stats-daemon/internal/model"
)

// Collect объединяет данные о блоках и inode в одну карту ФС.
func (c *Collector) Collect(ctx context.Context) (any, error) {
	spaceOut, err := c.exec.Run(ctx, "df", "-k")
	if err != nil {
		return nil, err
	}
	inodeOut, err := c.exec.Run(ctx, "df", "-i")
	if err != nil {
		return nil, err
	}

	fs := parseDF(spaceOut)
	mergeInodes(fs, inodeOut)
	return &model.FilesystemsStats{FS: fs}, nil
}

// parseDF разбирает вывод df -k.
func parseDF(out []byte) map[model.FilesystemKey]*model.FilesystemStats {
	result := make(map[model.FilesystemKey]*model.FilesystemStats)
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		if lineNo == 1 {
			continue
		}
		fields := strings.Fields(scanner.Text())
		if len(fields) < 6 {
			continue
		}
		usedKB, _ := strconv.ParseFloat(fields[2], 64)
		totalKB, _ := strconv.ParseFloat(fields[1], 64)
		usedPercent, _ := strconv.ParseFloat(strings.TrimSuffix(fields[4], "%"), 64)

		key := model.FilesystemKey{Name: fields[0], MountedOn: fields[5]}
		result[key] = &model.FilesystemStats{
			UsedMB:      usedKB / 1024,
			UsedPercent: usedPercent,
		}
		if totalKB > 0 && result[key].UsedPercent == 0 {
			result[key].UsedPercent = usedKB / totalKB * 100
		}
	}
	return result
}

// mergeInodes дополняет карту полями из df -i.
func mergeInodes(fs map[model.FilesystemKey]*model.FilesystemStats, inodeOut []byte) {
	scanner := bufio.NewScanner(strings.NewReader(string(inodeOut)))
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		if lineNo == 1 {
			continue
		}
		fields := strings.Fields(scanner.Text())
		if len(fields) < 6 {
			continue
		}
		used, _ := strconv.ParseFloat(fields[2], 64)
		total, _ := strconv.ParseFloat(fields[1], 64)
		usedPercent, _ := strconv.ParseFloat(strings.TrimSuffix(fields[4], "%"), 64)

		key := model.FilesystemKey{Name: fields[0], MountedOn: fields[5]}
		entry, ok := fs[key]
		if !ok {
			entry = &model.FilesystemStats{}
			fs[key] = entry
		}
		entry.UsedInodes = used
		entry.UsedInodesPercent = usedPercent
		if total > 0 && entry.UsedInodesPercent == 0 {
			entry.UsedInodesPercent = used / total * 100
		}
	}
}
