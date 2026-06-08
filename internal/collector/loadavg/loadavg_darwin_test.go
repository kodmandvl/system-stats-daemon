//go:build darwin

package loadavg

import (
	"context"
	"testing"

	"github.com/kodmandvl/system-stats-daemon/internal/collector"
	"github.com/kodmandvl/system-stats-daemon/internal/model"
)

type mockExec struct {
	out []byte
}

func (m *mockExec) Run(_ context.Context, _ string, _ ...string) ([]byte, error) {
	return m.out, nil
}

func TestParseUptime(t *testing.T) {
	c := New(&mockExec{out: []byte("15:28  up 51 days, load averages: 1.64 1.39 1.48\n")})
	res, err := c.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	la := res.(*model.LoadAvgStats)
	if la.LoadAvg1 != 1.64 {
		t.Fatalf("got %v", la.LoadAvg1)
	}
}

var _ collector.CommandExecutor = (*mockExec)(nil)
