package app

import (
	"log/slog"

	grpcapp "github.com/m1keee3/FinanceAnalyst/services/scanner/internal/app/grpc"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
) *App {

	// TODO db
	// TODO service
	grpcApp := grpcapp.New(log, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
