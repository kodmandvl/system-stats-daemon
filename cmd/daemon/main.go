// Демон системного мониторинга: gRPC server streaming + фоновый сбор метрик.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kodmandvl/system-stats-daemon/internal/app"
	"github.com/kodmandvl/system-stats-daemon/internal/collector"
	"github.com/kodmandvl/system-stats-daemon/internal/collector/cpu"
	"github.com/kodmandvl/system-stats-daemon/internal/collector/diskload"
	"github.com/kodmandvl/system-stats-daemon/internal/collector/filesystem"
	"github.com/kodmandvl/system-stats-daemon/internal/collector/loadavg"
	"github.com/kodmandvl/system-stats-daemon/internal/collector/network"
	"github.com/kodmandvl/system-stats-daemon/internal/config"
	grpcserver "github.com/kodmandvl/system-stats-daemon/internal/grpc/server"
	"github.com/kodmandvl/system-stats-daemon/internal/service/settings"
	"github.com/kodmandvl/system-stats-daemon/internal/service/stats"
	"github.com/kodmandvl/system-stats-daemon/internal/storage/memory"
)

func main() {
	// Порт gRPC и путь к JSON с включёнными подсистемами сбора.
	port := flag.String("port", "8080", "gRPC port")
	configPath := flag.String("config", "./configs/daemon.json", "path to JSON config")
	flag.Parse()

	cfg, err := config.LoadDaemon(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Слой сбора: команды ОС → парсинг → агрегация в один SystemStats.
	exec := &collector.OSExecutor{Timeout: cfg.CollectTimeout}
	collectors := buildCollectors(cfg, exec)
	aggregator := collector.NewAggregator(collectors...)

	// Память + усреднение; settings знает max(calc_period) от всех клиентов.
	repo := memory.NewRepository(memory.NewAverager(cfg.Metrics))
	settingsSvc := settings.New()
	statsSvc := stats.New(repo, aggregator, settingsSvc, time.Duration(cfg.CollectIntervalSec)*time.Second)
	application := app.New(statsSvc, settingsSvc)
	grpcSrv := grpcserver.New(application, *port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Фоновый тикер: раз в collect_interval_sec опрашивает систему.
	go statsSvc.Start(ctx)

	// gRPC server streaming — в отдельной goroutine.
	go func() {
		log.Printf("gRPC listening on :%s", *port)
		if err := grpcSrv.ListenAndServe(); err != nil {
			log.Printf("grpc stopped: %v", err)
			cancel()
		}
	}()

	// Graceful shutdown по SIGINT/SIGTERM.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	grpcSrv.Stop(shutdownCtx)
	cancel()
}

// buildCollectors собирает список коллекторов согласно флагам в configs/daemon.json.
func buildCollectors(cfg *config.DaemonConfig, exec collector.CommandExecutor) []collector.MetricCollector {
	var list []collector.MetricCollector
	if cfg.Metrics.LoadAvg {
		list = append(list, loadavg.New(exec))
	}
	if cfg.Metrics.CPU {
		list = append(list, cpu.New(exec))
	}
	if cfg.Metrics.DisksLoad {
		list = append(list, diskload.New(exec))
	}
	if cfg.Metrics.Filesystem {
		list = append(list, filesystem.New(exec))
	}
	if cfg.Metrics.Network {
		list = append(list, network.New(exec))
	}
	return list
}
