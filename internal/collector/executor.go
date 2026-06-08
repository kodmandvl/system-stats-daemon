// Выполнение внешних команд — см. package collector (collector.go).
package collector

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// CommandExecutor — обёртка над os/exec для тестирования через моки.
type CommandExecutor interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

// OSExecutor выполняет системные команды (top, df, ss и т.д.).
type OSExecutor struct {
	Timeout time.Duration
}

// Run запускает команду с таймаутом CollectTimeout из конфига.
func (e *OSExecutor) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, e.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("%s %v: %w", name, args, err)
	}
	return out, nil
}
