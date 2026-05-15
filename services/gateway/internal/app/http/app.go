package httpapp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

type App struct {
	log    *slog.Logger
	server *http.Server
}

func New(log *slog.Logger, handler http.Handler, port int) *App {
	return &App{
		log: log,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: handler,
		},
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "http.App.Run"
	a.log.Info("starting http server", slog.String("op", op), slog.String("addr", a.server.Addr))
	if err := a.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *App) Stop() {
	const op = "http.App.Stop"
	a.log.Info("stopping http server", slog.String("op", op))
	if err := a.server.Shutdown(context.Background()); err != nil {
		a.log.Warn("failed to shutdown http server", slog.String("op", op))
	}
}
