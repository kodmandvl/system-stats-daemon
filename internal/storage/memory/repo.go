// Package memory хранит сырые снимки в RAM и отдаёт усреднённые за окно M секунд.
package memory

import (
	"sync"

	"github.com/kodmandvl/system-stats-daemon/internal/model"
)

// Repository хранит последние M секунд сырых снимков в памяти.
type Repository struct {
	mu    sync.RWMutex
	cache []*model.SystemStats // снимки раз в секунду, не дольше max(calc_period)
	avg   *Averager
}

// NewRepository создаёт in-memory хранилище с заданной стратегией усреднения.
func NewRepository(avg *Averager) *Repository {
	return &Repository{
		cache: make([]*model.SystemStats, 0),
		avg:   avg,
	}
}

// Update добавляет новый снимок и обрезает кэш до periodToKeep элементов.
func (r *Repository) Update(stats *model.SystemStats, periodToKeep uint32) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cache = append(r.cache, cloneStats(stats))
	if len(r.cache) > int(periodToKeep) {
		r.cache = r.cache[len(r.cache)-int(periodToKeep):]
	}
}

// GetAvg возвращает усреднённый снимок за последние period секунд.
func (r *Repository) GetAvg(period uint32) (*model.SystemStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if period == 0 {
		return nil, model.ErrPeriodNotValid
	}
	if len(r.cache) == 0 {
		return emptyStats(), nil
	}

	start := 0
	if len(r.cache) > int(period) {
		start = len(r.cache) - int(period)
	}
	window := r.cache[start:]
	return r.avg.Average(window, float64(len(window))), nil
}

// emptyStats — заглушка, пока клиент подключился, но сбор ещё не накопил данные.
func emptyStats() *model.SystemStats {
	return &model.SystemStats{
		LoadAvg:     &model.LoadAvgStats{},
		CPU:         &model.CPUStats{},
		DisksLoad:   &model.DisksLoadStats{Disks: map[string]*model.DiskLoad{}},
		Filesystems: &model.FilesystemsStats{FS: map[model.FilesystemKey]*model.FilesystemStats{}},
		Network:     &model.NetworkStats{},
	}
}

func cloneStats(s *model.SystemStats) *model.SystemStats {
	if s == nil {
		return emptyStats()
	}
	// Поверхностное копирование достаточно: при Update создаются новые объекты коллектора.
	return s
}
