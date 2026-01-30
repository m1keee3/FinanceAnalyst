package watcher

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/clients/watcher/mapper"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/models/watcher"
	watcherv1 "github.com/m1keee3/FinanceAnalyst/services/gateway/proto-gen/watcher/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"time"
)

type Client struct {
	api     watcherv1.WatcherClient
	conn    *grpc.ClientConn
	timeout time.Duration
}

func New(log *slog.Logger, addr string, timeout time.Duration) (*Client, error) {

	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadReceived,
			logging.PayloadSent,
		),
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			logging.UnaryClientInterceptor(InterceptorLogger(log), loggingOpts...),
		),
	}

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &Client{
		api:     watcherv1.NewWatcherClient(conn),
		conn:    conn,
		timeout: timeout,
	}, nil
}

func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func (c *Client) GetStats(ctx context.Context) ([]*watcher.SegmentStats, error) {
	const op = "watcher.Client.GetStats"

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	resp, err := c.api.GetStats(ctx, &watcherv1.GetStatsRequest{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	res := mapper.FromStatsResponse(resp)
	return res, err
}
