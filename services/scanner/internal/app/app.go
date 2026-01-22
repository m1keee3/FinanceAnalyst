package app

import (
	"github.com/m1keee3/FinanceAnalyst/pkg/logger/sl"
	"log/slog"
	"time"

	grpcapp "github.com/m1keee3/FinanceAnalyst/services/scanner/internal/app/grpc"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/cache/redis"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/scanner"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/scanner/candle"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/scanner/chart"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/scanner/stats"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/lib/fetcher/moex"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpcapp.App
	cache      *redis.Cache
}

func New(
	log *slog.Logger,
	cacheAddr string,
	cachePassword string,
	cacheDB int,
	grpcPort int,
	requestTimeout time.Duration,
) *App {

	cache := redis.NewCache(cacheAddr, cachePassword, cacheDB)

	fetcher := moex.NewFetcher()
	scannerService := scanner.NewService(
		log,
		candle.NewScanner(log, fetcher),
		chart.NewScanner(log, fetcher),
		stats.NewComputer(fetcher),
		cache,
	)

	grpcApp := grpcapp.New(log, scannerService, grpcPort, requestTimeout)

	return &App{
		log:        log,
		gRPCServer: grpcApp,
		cache:      cache,
	}
}

func (a *App) Run() {
	const op = "app.Run"

	a.log.Info("starting app", slog.String("op", op))
	a.gRPCServer.MustRun()
}

func (a *App) Stop() {
	const op = "app.Stop"

	log := a.log.With(slog.String("op", op))

	a.gRPCServer.Stop()
	err := a.cache.Close()
	if err != nil {
		log.Warn("failed to close cache", sl.Err(err))
	}

	log.Info("app stopped")
}
