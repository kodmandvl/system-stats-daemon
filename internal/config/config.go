// Package config загружает настройки демона из JSON-файла.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Metrics — флаги включения отдельных коллекторов (2 балла в ТЗ).
type Metrics struct {
	LoadAvg    bool `json:"loadAvg"`
	CPU        bool `json:"cpu"`
	DisksLoad  bool `json:"disksLoad"`
	Filesystem bool `json:"filesystem"`
	Network    bool `json:"network"`
}

// DaemonConfig — конфигурация демона из JSON-файла.
type DaemonConfig struct {
	CollectIntervalSec int           `json:"collectIntervalSec"` // период опроса системы (сек)
	CollectTimeout     time.Duration `json:"-"`                  // заполняется из CollectTimeoutSec
	CollectTimeoutSec  int           `json:"collectTimeoutSec"`  // таймаут одной команды (top, ss, …)
	Metrics            Metrics       `json:"metrics"`
}

// ClientConfig — настройки CLI-клиента.
type ClientConfig struct {
	Address    string `json:"address"`
	SendPeriod uint32 `json:"sendPeriod"`
	CalcPeriod uint32 `json:"calcPeriod"`
	// Section — какой раздел выводить: network, cpu, load, disks, fs, all.
	Section string `json:"section"`
}

// LoadDaemon читает JSON-конфиг демона.
func LoadDaemon(path string) (*DaemonConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg DaemonConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.CollectIntervalSec <= 0 {
		cfg.CollectIntervalSec = 1
	}
	if cfg.CollectTimeoutSec <= 0 {
		cfg.CollectTimeoutSec = 5
	}
	cfg.CollectTimeout = time.Duration(cfg.CollectTimeoutSec) * time.Second

	// По умолчанию включаем все метрики, если секция metrics пустая в файле.
	if !cfg.Metrics.LoadAvg && !cfg.Metrics.CPU && !cfg.Metrics.DisksLoad &&
		!cfg.Metrics.Filesystem && !cfg.Metrics.Network {
		cfg.Metrics = Metrics{
			LoadAvg: true, CPU: true, DisksLoad: true, Filesystem: true, Network: true,
		}
	}

	return &cfg, nil
}

// LoadClient читает JSON-конфиг клиента.
func LoadClient(path string) (*ClientConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg ClientConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.SendPeriod == 0 {
		cfg.SendPeriod = 5
	}
	if cfg.CalcPeriod == 0 {
		cfg.CalcPeriod = 15
	}
	if cfg.Section == "" {
		cfg.Section = "network"
	}

	return &cfg, nil
}
