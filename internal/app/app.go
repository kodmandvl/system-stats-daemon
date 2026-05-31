// Package app реализует бизнес-логику server-side streaming (N и M из ТЗ).
package app

import (
	"context"
	"time"

	"github.com/kodmandvl/system-stats-daemon/internal/model"
	"github.com/kodmandvl/system-stats-daemon/internal/service/settings"
)

// StatsReader отдаёт усреднённые снапшоты.
type StatsReader interface {
	Snapshot(calcPeriod uint32) (*model.SystemStats, error)
}

// Application реализует логику server-side streaming из ТЗ.
type Application struct {
	stats    StatsReader
	settings *settings.Service
}

// New связывает чтение снапшотов с учётом активных клиентов.
func New(stats StatsReader, settings *settings.Service) *Application {
	return &Application{stats: stats, settings: settings}
}

// ObserveSystemStats реализует тайминг из ТЗ (пример N=5, M=15):
//   - молчит первые M секунд (15 с);
//   - в 15 с — снапшот за 0–15 с, в 20 с — за 5–20 с, в 25 с — за 10–25 с и т.д.
//
// Технически: ждём (M−N), затем шлём по тикеру с периодом N (первый тик ровно на M-й секунде).
func (a *Application) ObserveSystemStats(
	ctx context.Context,
	send func(*model.SystemStats) error,
	sendPeriod, calcPeriod uint32,
) error {
	if sendPeriod == 0 || calcPeriod == 0 || sendPeriod > calcPeriod {
		return model.ErrPeriodNotValid
	}

	a.settings.Add(calcPeriod)
	defer a.settings.Remove(calcPeriod)

	// До первой отправки проходит M секунд: (M−N) ожидание + N до первого тика.
	pauseBeforeFirstTick := time.Duration(calcPeriod-sendPeriod) * time.Second
	if pauseBeforeFirstTick > 0 {
		timer := time.NewTimer(pauseBeforeFirstTick)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}

	ticker := time.NewTicker(time.Duration(sendPeriod) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			snap, err := a.stats.Snapshot(calcPeriod)
			if err != nil {
				return err
			}
			if err := send(snap); err != nil {
				return err
			}
		}
	}
}
