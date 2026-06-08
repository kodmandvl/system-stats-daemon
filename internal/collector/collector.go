// Package collector опрашивает подсистемы ОС и объединяет результаты.
package collector

import (
	"context"
	"sync"

	"github.com/kodmandvl/system-stats-daemon/internal/model"
)

// MetricCollector — один источник метрик (load, cpu, ...).
type MetricCollector interface {
	Name() string
	Collect(ctx context.Context) (any, error)
}

// Aggregator параллельно опрашивает все включённые коллекторы.
type Aggregator struct {
	collectors []MetricCollector
}

// NewAggregator создаёт агрегатор с заданным списком коллекторов.
func NewAggregator(collectors ...MetricCollector) *Aggregator {
	active := make([]MetricCollector, 0, len(collectors))
	for _, c := range collectors {
		if c != nil {
			active = append(active, c)
		}
	}
	return &Aggregator{collectors: active}
}

// Collect собирает снапшот, запуская коллекторы конкурентно.
func (a *Aggregator) Collect(ctx context.Context) (*model.SystemStats, error) {
	result := &model.SystemStats{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	errCh := make(chan error, len(a.collectors))

	for _, col := range a.collectors {
		wg.Add(1)
		go func(c MetricCollector) {
			defer wg.Done()
			data, err := c.Collect(ctx)
			if err != nil {
				errCh <- err
				return
			}
			mu.Lock()
			mergeStats(result, data)
			mu.Unlock()
		}(col)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// mergeStats раскладывает результат Collect по полям SystemStats.
func mergeStats(dst *model.SystemStats, data any) {
	switch v := data.(type) {
	case *model.LoadAvgStats:
		dst.LoadAvg = v
	case *model.CPUStats:
		dst.CPU = v
	case *model.DisksLoadStats:
		dst.DisksLoad = v
	case *model.FilesystemsStats:
		dst.Filesystems = v
	case *model.NetworkStats:
		dst.Network = v
	}
}
