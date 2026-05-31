package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kodmandvl/system-stats-daemon/internal/model"
	"github.com/kodmandvl/system-stats-daemon/internal/service/settings"
)

type mockReader struct {
	calls int
}

func (m *mockReader) Snapshot(uint32) (*model.SystemStats, error) {
	m.calls++
	return &model.SystemStats{LoadAvg: &model.LoadAvgStats{LoadAvg1: 1}}, nil
}

func TestObserveSystemStatsPeriodValidation(t *testing.T) {
	app := New(&mockReader{}, settings.New())
	err := app.ObserveSystemStats(context.Background(), func(*model.SystemStats) error { return nil }, 10, 5)
	if !errors.Is(err, model.ErrPeriodNotValid) {
		t.Fatalf("expected period error, got %v", err)
	}
}

func TestObserveSystemStatsSendsSnapshot(t *testing.T) {
	reader := &mockReader{}
	application := New(reader, settings.New())
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// При N=M=1 первая отправка — на 1-й секунде (после первого тика).
		time.Sleep(1100 * time.Millisecond)
		cancel()
	}()

	sent := 0
	err := application.ObserveSystemStats(ctx, func(*model.SystemStats) error {
		sent++
		return nil
	}, 1, 1)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected err: %v", err)
	}
	if sent < 1 {
		t.Fatalf("expected at least 1 snapshot, got %d", sent)
	}
}
