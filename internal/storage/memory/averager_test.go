package memory

import (
	"testing"

	"github.com/kodmandvl/system-stats-daemon/internal/config"
	"github.com/kodmandvl/system-stats-daemon/internal/model"
)

func TestAveragerAverage(t *testing.T) {
	avg := NewAverager(config.Metrics{LoadAvg: true, CPU: true})
	window := []*model.SystemStats{
		{
			LoadAvg: &model.LoadAvgStats{LoadAvg1: 2, LoadAvg5: 4, LoadAvg15: 6},
			CPU:     &model.CPUStats{UserMode: 10, SystemMode: 20, Idle: 70},
		},
		{
			LoadAvg: &model.LoadAvgStats{LoadAvg1: 4, LoadAvg5: 6, LoadAvg15: 8},
			CPU:     &model.CPUStats{UserMode: 30, SystemMode: 40, Idle: 30},
		},
	}
	result := avg.Average(window, 2)
	if result.LoadAvg.LoadAvg1 != 3 {
		t.Fatalf("load1: got %v want 3", result.LoadAvg.LoadAvg1)
	}
	if result.CPU.UserMode != 20 {
		t.Fatalf("cpu user: got %v want 20", result.CPU.UserMode)
	}
}
