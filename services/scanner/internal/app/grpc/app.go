package grpcapp

import (
	"fmt"
	"log/slog"
	"net"

	scannergrpc "github.com/m1keee3/FinanceAnalyst/services/scanner/internal/grpc"
	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	grpcServer *grpc.Server
	port       int
}

func New(
	log *slog.Logger,
	port int,
) *App {
	grpcServer := grpc.NewServer()

	scannergrpc.Register(grpcServer)

	return &App{
		log:        log,
		grpcServer: grpcServer,
		port:       port,
	}
}

// MustRun runs gRPC server and panics if any error occurs.
func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	log := a.log.With(slog.String("op", op))

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("starting grpc server", slog.String("addr", l.Addr().String()))

	if err := a.grpcServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping grpc server on port %d", a.port)

	a.grpcServer.GracefulStop()
}
