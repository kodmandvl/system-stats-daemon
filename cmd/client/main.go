// Простой gRPC-клиент: подключается к демону и печатает выбранную секцию в STDOUT.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"

	"github.com/kodmandvl/system-stats-daemon/internal/config"
	"github.com/kodmandvl/system-stats-daemon/internal/grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run подключается к демону и читает поток ObserveSystemStats до отмены контекста.
func run() error {
	addr := flag.String("addr", "", "gRPC address host:port")
	configPath := flag.String("config", "./configs/client.json", "client JSON config")
	flag.Parse()

	cfg, err := config.LoadClient(*configPath)
	if err != nil {
		return err
	}
	if *addr != "" {
		cfg.Address = *addr
	}
	if cfg.Address == "" {
		cfg.Address = "localhost:8080"
	}

	ctx, cancel := signalContext()
	defer cancel()

	conn, err := grpc.NewClient(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewSystemStatsServiceClient(conn)
	// N и M задают период отправки и окно усреднения (см. ТЗ).
	stream, err := client.ObserveSystemStats(ctx, &pb.SystemStatsRequest{
		SendPeriod: cfg.SendPeriod,
		CalcPeriod: cfg.CalcPeriod,
	})
	if err != nil {
		return err
	}

	for {
		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		printSection(os.Stdout, cfg.Section, msg)
		fmt.Println("---")
	}
}

// signalContext отменяет контекст при Ctrl+C / SIGTERM.
func signalContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ch
		cancel()
	}()
	return ctx, cancel
}

// printSection выводит одну подсистему в виде таблицы (tabwriter).
// section: network | cpu | load | default — полный дамп.
func printSection(w io.Writer, section string, msg *pb.SystemStats) {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	defer tw.Flush()

	switch section {
	case "network":
		if msg.Network == nil {
			fmt.Fprintln(w, "no network stats")
			return
		}
		fmt.Fprintln(tw, "PROTOCOL\tBYTES\t%")
		for _, p := range msg.Network.ProtocolTalkers {
			fmt.Fprintf(tw, "%s\t%d\t%.2f\n", p.Protocol, p.Bytes, p.Percent)
		}
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "SRC\tDST\tPROTO\tBPS")
		for _, f := range msg.Network.FlowTalkers {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%.2f\n", f.Source, f.Destination, f.Protocol, f.Bps)
		}
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "STATE\tCOUNT")
		for _, s := range msg.Network.TcpStates {
			fmt.Fprintf(tw, "%s\t%d\n", s.State, s.Count)
		}
	case "cpu":
		if msg.Cpu != nil {
			fmt.Fprintf(tw, "user\t%.2f\nsystem\t%.2f\nidle\t%.2f\n",
				msg.Cpu.UserMode, msg.Cpu.SystemMode, msg.Cpu.Idle)
		}
	case "load":
		if msg.LoadAvg != nil {
			fmt.Fprintf(tw, "1m\t%.2f\n5m\t%.2f\n15m\t%.2f\n",
				msg.LoadAvg.LoadAvg1, msg.LoadAvg.LoadAvg5, msg.LoadAvg.LoadAvg15)
		}
	default:
		fmt.Fprintf(w, "%+v\n", msg)
	}
}
