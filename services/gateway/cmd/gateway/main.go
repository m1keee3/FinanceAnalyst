package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/m1keee3/FinanceAnalyst/pkg/logger/handlers/slogpretty"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/app"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/config"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	application := app.New(log, cfg)

	go application.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	sign := <-stop

	log.Info("stopping app", slog.String("signal", sign.String()))

	application.Stop()
}

func setupLogger(env string) *slog.Logger {
	switch env {
	case envLocal:
		return setupPrettySlog()
	case envDev:
		return slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	default:
		return slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}
	return slog.New(opts.NewPrettyHandler(os.Stdout))
}
