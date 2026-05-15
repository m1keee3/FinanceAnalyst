package app

import (
	"log/slog"

	"github.com/m1keee3/FinanceAnalyst/pkg/logger/sl"
	httpapp "github.com/m1keee3/FinanceAnalyst/services/gateway/internal/app/http"
	scannerclient "github.com/m1keee3/FinanceAnalyst/services/gateway/internal/clients/scanner"
	watcherclient "github.com/m1keee3/FinanceAnalyst/services/gateway/internal/clients/watcher"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/config"
	httpserver "github.com/m1keee3/FinanceAnalyst/services/gateway/internal/http-server"
	scannerhandler "github.com/m1keee3/FinanceAnalyst/services/gateway/internal/http-server/handlers/scanner"
	watcherhandler "github.com/m1keee3/FinanceAnalyst/services/gateway/internal/http-server/handlers/watcher"
)

type App struct {
	log           *slog.Logger
	httpServer    *httpapp.App
	scannerClient *scannerclient.Client
	watcherClient *watcherclient.Client
}

func New(log *slog.Logger, cfg *config.Config) *App {
	scannerClient, err := scannerclient.New(log, cfg.Scanner.Address, cfg.Scanner.Timeout)
	if err != nil {
		log.Error("failed to connect to scanner", sl.Err(err))
	}

	watcherClient, err := watcherclient.New(log, cfg.Watcher.Address, cfg.Watcher.Timeout)
	if err != nil {
		log.Error("failed to connect to watcher", sl.Err(err))
	}

	scannerH := scannerhandler.New(log, scannerClient)
	watcherH := watcherhandler.New(log, watcherClient)

	router := httpserver.NewRouter(log, scannerH, watcherH)
	httpApp := httpapp.New(log, router, cfg.HTTP.Port)

	return &App{
		log:           log,
		httpServer:    httpApp,
		scannerClient: scannerClient,
		watcherClient: watcherClient,
	}
}

func (a *App) Run() {
	a.httpServer.MustRun()
}

func (a *App) Stop() {
	const op = "app.Stop"

	log := a.log.With(slog.String("op", op))

	a.httpServer.Stop()

	if a.scannerClient != nil {
		if err := a.scannerClient.Close(); err != nil {
			log.Warn("failed to close scanner client", sl.Err(err))
		}
	}

	if a.watcherClient != nil {
		if err := a.watcherClient.Close(); err != nil {
			log.Warn("failed to close watcher client", sl.Err(err))
		}
	}

	log.Info("app stopped")
}
