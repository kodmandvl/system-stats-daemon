// Package server — gRPC-обёртка над app.Application.
package server

import (
	"context"
	"fmt"
	"net"

	"github.com/kodmandvl/system-stats-daemon/internal/app"
	"github.com/kodmandvl/system-stats-daemon/internal/grpc/pb"
	"github.com/kodmandvl/system-stats-daemon/internal/model"
	"google.golang.org/grpc"
)

// GRPCServer — обёртка над gRPC.
type GRPCServer struct {
	pb.UnimplementedSystemStatsServiceServer
	app    *app.Application
	port   string
	server *grpc.Server
}

// New принимает порт без ведущего двоеточия (например "8080").
func New(application *app.Application, port string) *GRPCServer {
	return &GRPCServer{app: application, port: port}
}

// ObserveSystemStats — единственный RPC: поток SystemStats клиенту.
func (s *GRPCServer) ObserveSystemStats(
	req *pb.SystemStatsRequest,
	stream pb.SystemStatsService_ObserveSystemStatsServer,
) error {
	return s.app.ObserveSystemStats(stream.Context(), func(stats *model.SystemStats) error {
		return stream.Send(toProto(stats))
	}, req.SendPeriod, req.CalcPeriod)
}

// ListenAndServe блокируется до ошибки или Stop.
func (s *GRPCServer) ListenAndServe() error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	s.server = grpc.NewServer()
	pb.RegisterSystemStatsServiceServer(s.server, s)
	return s.server.Serve(lis)
}

// Stop завершает сервер: GracefulStop или принудительный Stop по таймауту ctx.
func (s *GRPCServer) Stop(ctx context.Context) {
	if s.server == nil {
		return
	}
	stopped := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(stopped)
	}()
	select {
	case <-ctx.Done():
		s.server.Stop()
	case <-stopped:
	}
}
