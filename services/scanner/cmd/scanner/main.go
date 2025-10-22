package main

import (
	"log/slog"
	"os"
	"os/signal"

	"github.com/m1keee3/FinanceAnalyst/pkg/logger/handlers/slogpretty"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/app"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/config"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	application := app.New(log, cfg.Grpc.Port)

	go application.GRPCServer.MustRun()

	// TODO db application

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	sign := <-stop

	log.Info("scanner stopped", slog.Any("signal", sign.String()))

	application.GRPCServer.Stop()

	log.Info("scanner stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
