//go:build integration

// Интеграционные тесты gRPC-потока (in-memory bufconn, без реальной сети).
package integration

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/kodmandvl/system-stats-daemon/internal/app"
	"github.com/kodmandvl/system-stats-daemon/internal/collector"
	"github.com/kodmandvl/system-stats-daemon/internal/config"
	grpcserver "github.com/kodmandvl/system-stats-daemon/internal/grpc/server"
	"github.com/kodmandvl/system-stats-daemon/internal/grpc/pb"
	"github.com/kodmandvl/system-stats-daemon/internal/service/settings"
	"github.com/kodmandvl/system-stats-daemon/internal/service/stats"
	"github.com/kodmandvl/system-stats-daemon/internal/storage/memory"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

// startTestServer поднимает gRPC в памяти (bufconn) для теста без TCP.
func startTestServer(t *testing.T) (*grpc.ClientConn, func()) {
	t.Helper()
	lis := bufconn.Listen(bufSize)

	cfg := config.Metrics{LoadAvg: true, CPU: false, DisksLoad: false, Filesystem: false, Network: false}
	repo := memory.NewRepository(memory.NewAverager(cfg))
	settingsSvc := settings.New()
	agg := collector.NewAggregator() // пустой коллектор — проверяем только поток
	statsSvc := stats.New(repo, agg, settingsSvc, time.Second)
	application := app.New(statsSvc, settingsSvc)
	srv := grpcserver.New(application, "0")

	ctx, cancel := context.WithCancel(context.Background())
	go statsSvc.Start(ctx)

	gs := grpc.NewServer()
	pb.RegisterSystemStatsServiceServer(gs, srv)

	go func() {
		if err := gs.Serve(lis); err != nil {
			t.Logf("serve: %v", err)
		}
	}()

	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	cleanup := func() {
		cancel()
		conn.Close()
		gs.Stop()
	}
	return conn, cleanup
}

// TestStreamReceivesSnapshots проверяет, что клиент получает несколько сообщений потока.
func TestStreamReceivesSnapshots(t *testing.T) {
	conn, cleanup := startTestServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	client := pb.NewSystemStatsServiceClient(conn)
	stream, err := client.ObserveSystemStats(ctx, &pb.SystemStatsRequest{SendPeriod: 1, CalcPeriod: 2})
	if err != nil {
		t.Fatal(err)
	}

	received := 0
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		received++
		if received >= 2 {
			cancel()
			break
		}
	}
	if received < 2 {
		t.Fatalf("expected >=2 messages, got %d", received)
	}
}
