package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/m1keee3/FinanceAnalyst/services/scanner/internal/app/grpc"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/cache/redis"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/scanner"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/scanner/candle"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/scanner/chart"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/scanner/stats"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/lib/fetcher/moex"
	_default "github.com/m1keee3/FinanceAnalyst/services/scanner/lib/ttl_resolver/default"
)

type App struct {
	GRPCServer *grpcapp.App
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
		_default.NewTTLResolver(time.Minute*5, time.Hour*3),
	)

	grpcApp := grpcapp.New(log, scannerService, grpcPort, requestTimeout)

	return &App{
		GRPCServer: grpcApp,
	}
}
