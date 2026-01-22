package app

import (
	"fmt"
	"github.com/m1keee3/FinanceAnalyst/pkg/logger/sl"
	"github.com/m1keee3/FinanceAnalyst/services/watcher/internal/app/grpc"
	"github.com/m1keee3/FinanceAnalyst/services/watcher/internal/clients/scanner"
	"github.com/m1keee3/FinanceAnalyst/services/watcher/internal/config"
	"github.com/m1keee3/FinanceAnalyst/services/watcher/internal/kafka"
	"github.com/m1keee3/FinanceAnalyst/services/watcher/internal/scheduler"
	"github.com/m1keee3/FinanceAnalyst/services/watcher/internal/services/watcher"
	"github.com/m1keee3/FinanceAnalyst/services/watcher/internal/storage/postgres"
	"log/slog"
	"net/url"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.App
	storage    *postgres.Storage
	kafka      *kafka.Producer
	scheduler  *scheduler.Scheduler
}

func New(log *slog.Logger, cfg *config.Config) *App {
	const op = "app.New"

	producer := kafka.New(cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.Timeout)
	log.Info("connected to kafka", slog.String("op", op), slog.Any("brokers", cfg.Kafka.Brokers), slog.String("topic", cfg.Kafka.Topic))

	scannerClient, err := scanner.New(log, cfg.GrpcClient.Address, cfg.GrpcClient.Timeout)
	if err != nil {
		log.Error("failed to connect to scanner", sl.Err(err))
	} else {
		log.Info("connected to scanner client", slog.String("op", op), slog.Any("address", cfg.GrpcClient.Address))
	}

	storage, err := postgres.New(buildDSN(&cfg.Db))
	if err != nil {
		log.Error("failed to connect to postgres", sl.Err(err))
	} else {
		log.Info("connected to postgres", slog.String("op", op))
	}

	w := watcher.New(log, &cfg.App, producer, storage, scannerClient)

	s, err := scheduler.New(cfg.Scheduler.Cron, cfg.Scheduler.Timezone, w)
	if err != nil {
		log.Error("failed to create scheduler", sl.Err(err))
	}

	grpcApp := grpc.New(log, w, cfg.Grpc.Port, cfg.Grpc.RequestTimeout)

	return &App{
		log:        log,
		gRPCServer: grpcApp,
		storage:    storage,
		kafka:      producer,
		scheduler:  s,
	}
}

func (a *App) Run() {
	const op = "app.Run"

	a.log.Info("starting app", slog.String("op", op))

	a.scheduler.Run()
	a.gRPCServer.MustRun()
}

func (a *App) Stop() {
	const op = "app.Stop"

	log := a.log.With(slog.String("op", op))

	a.scheduler.Stop()

	a.gRPCServer.Stop()

	if err := a.kafka.Close(); err != nil {
		log.Warn("failed to close kafka", sl.Err(err))
	}

	if err := a.storage.Close(); err != nil {
		log.Warn("failed to close storage", sl.Err(err))
	}
}

func buildDSN(cfg *config.DbConfig) string {
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:   cfg.DBName,
	}

	q := u.Query()
	q.Set("sslmode", cfg.SSLMode)
	u.RawQuery = q.Encode()

	return u.String()
}
