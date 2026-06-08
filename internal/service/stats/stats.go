// Package stats запускает периодический сбор и записывает снимки в репозиторий.
package stats

import (
	"context"
	"time"

	"github.com/kodmandvl/system-stats-daemon/internal/model"
)

// Collector собирает сырые метрики.
type Collector interface {
	Collect(ctx context.Context) (*model.SystemStats, error)
}

// Repository хранит историю снимков.
type Repository interface {
	Update(stats *model.SystemStats, periodToKeep uint32)
	GetAvg(period uint32) (*model.SystemStats, error)
}

// Settings хранит активные окна клиентов.
type Settings interface {
	Max() (uint32, bool)
}

// Service периодически собирает статистику и кладёт в репозиторий.
type Service struct {
	repo      Repository
	collector Collector
	settings  Settings
	interval  time.Duration
}

// New создаёт сервис с интервалом опроса (обычно 1 с).
func New(repo Repository, collector Collector, settings Settings, interval time.Duration) *Service {
	return &Service{
		repo:      repo,
		collector: collector,
		settings:  settings,
		interval:  interval,
	}
}

// Start крутит ticker до отмены ctx; без подключённых клиентов сбор не выполняется.
func (s *Service) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.collectOnce(ctx)
		}
	}
}

// collectOnce опрашивает систему, если есть хотя бы один gRPC-клиент.
func (s *Service) collectOnce(ctx context.Context) {
	period, ok := s.settings.Max()
	if !ok {
		return
	}
	stats, err := s.collector.Collect(ctx)
	if err != nil {
		return
	}
	s.repo.Update(stats, period)
}

// Snapshot отдаёт усреднённый снимок за calcPeriod — вызывается из gRPC-потока.
func (s *Service) Snapshot(calcPeriod uint32) (*model.SystemStats, error) {
	return s.repo.GetAvg(calcPeriod)
}
