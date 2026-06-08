// Усреднение и слияние снимков — см. package memory (repo.go).
package memory

import (
	"sort"

	"github.com/kodmandvl/system-stats-daemon/internal/config"
	"github.com/kodmandvl/system-stats-daemon/internal/model"
)

// Averager агрегирует окно снимков с учётом включённых метрик.
type Averager struct {
	enabled config.Metrics
}

// NewAverager учитывает только включённые в конфиге подсистемы.
func NewAverager(enabled config.Metrics) *Averager {
	return &Averager{enabled: enabled}
}

// Average суммирует окно снимков и делит на period (арифметическое среднее).
func (a *Averager) Average(window []*model.SystemStats, period float64) *model.SystemStats {
	if period <= 0 {
		period = 1
	}
	result := emptyStats()

	for _, snap := range window {
		if a.enabled.LoadAvg && snap.LoadAvg != nil {
			result.LoadAvg.LoadAvg1 += snap.LoadAvg.LoadAvg1
			result.LoadAvg.LoadAvg5 += snap.LoadAvg.LoadAvg5
			result.LoadAvg.LoadAvg15 += snap.LoadAvg.LoadAvg15
		}
		if a.enabled.CPU && snap.CPU != nil {
			result.CPU.UserMode += snap.CPU.UserMode
			result.CPU.SystemMode += snap.CPU.SystemMode
			result.CPU.Idle += snap.CPU.Idle
		}
		if a.enabled.DisksLoad && snap.DisksLoad != nil {
			mergeDisks(result.DisksLoad.Disks, snap.DisksLoad.Disks)
		}
		if a.enabled.Filesystem && snap.Filesystems != nil {
			mergeFilesystems(result.Filesystems.FS, snap.Filesystems.FS)
		}
		if a.enabled.Network && snap.Network != nil {
			mergeNetwork(result.Network, snap.Network)
		}
	}

	if a.enabled.LoadAvg {
		result.LoadAvg.LoadAvg1 /= period
		result.LoadAvg.LoadAvg5 /= period
		result.LoadAvg.LoadAvg15 /= period
	}
	if a.enabled.CPU {
		result.CPU.UserMode /= period
		result.CPU.SystemMode /= period
		result.CPU.Idle /= period
	}
	if a.enabled.DisksLoad {
		for _, d := range result.DisksLoad.Disks {
			d.Tps /= period
			d.Kbs /= period
		}
	}
	if a.enabled.Filesystem {
		for _, fs := range result.Filesystems.FS {
			fs.UsedMB /= period
			fs.UsedPercent /= period
			fs.UsedInodes /= period
			fs.UsedInodesPercent /= period
		}
	}
	if a.enabled.Network {
		for i := range result.Network.FlowTalkers {
			result.Network.FlowTalkers[i].Bps /= period
		}
		normalizeProtocolPercents(result.Network.ProtocolTalkers)
		sort.Slice(result.Network.FlowTalkers, func(i, j int) bool {
			return result.Network.FlowTalkers[i].Bps > result.Network.FlowTalkers[j].Bps
		})
	}

	return result
}

// mergeDisks суммирует метрики по имени устройства.
func mergeDisks(dst, src map[string]*model.DiskLoad) {
	for name, load := range src {
		if dst[name] == nil {
			dst[name] = &model.DiskLoad{}
		}
		dst[name].Tps += load.Tps
		dst[name].Kbs += load.Kbs
	}
}

// mergeFilesystems суммирует занятость по каждой точке монтирования.
func mergeFilesystems(dst, src map[model.FilesystemKey]*model.FilesystemStats) {
	for k, v := range src {
		if dst[k] == nil {
			dst[k] = &model.FilesystemStats{}
		}
		dst[k].UsedMB += v.UsedMB
		dst[k].UsedPercent += v.UsedPercent
		dst[k].UsedInodes += v.UsedInodes
		dst[k].UsedInodesPercent += v.UsedInodesPercent
	}
}

// mergeNetwork объединяет сетевые срезы; для сокетов/states берётся последний снимок.
func mergeNetwork(dst, src *model.NetworkStats) {
	dst.ProtocolTalkers = append(dst.ProtocolTalkers, src.ProtocolTalkers...)
	dst.FlowTalkers = append(dst.FlowTalkers, src.FlowTalkers...)
	dst.ListeningSockets = src.ListeningSockets
	dst.TCPStates = src.TCPStates
}

// normalizeProtocolPercents пересчитывает доли % от суммы байт за окно.
func normalizeProtocolPercents(talkers []model.ProtocolTalker) {
	var total uint64
	for _, t := range talkers {
		total += t.Bytes
	}
	for i := range talkers {
		if total > 0 {
			talkers[i].Percent = float64(talkers[i].Bytes) / float64(total) * 100
		}
	}
	sort.Slice(talkers, func(i, j int) bool { return talkers[i].Percent > talkers[j].Percent })
}
