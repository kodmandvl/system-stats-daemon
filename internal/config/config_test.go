package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDaemonDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "daemon.json")
	if err := os.WriteFile(path, []byte(`{"metrics":{"loadAvg":true}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadDaemon(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.CollectIntervalSec != 1 {
		t.Fatalf("interval: %d", cfg.CollectIntervalSec)
	}
}
