package scanner

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/clients/scanner/mapper"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/models/scanner"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/models/scanner/candle"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/models/scanner/chart"
	scannerv1 "github.com/m1keee3/FinanceAnalyst/services/gateway/proto-gen/scanner/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"time"
)

type Client struct {
	api     scannerv1.ScannerClient
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
		api:     scannerv1.NewScannerClient(conn),
		conn:    conn,
		timeout: timeout,
	}, nil
}

func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func (c *Client) FindCandleMatches(ctx context.Context, req *candle.ScanRequest) ([]*scanner.ChartSegment, error) {
	const op = "scanner.Client.FindCandleMatches"

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	resp, err := c.api.FindCandleMatches(ctx, mapper.ToCandleScanRequest(req))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	res := mapper.FromCandleMatchesResponse(resp)
	return res, nil
}

func (c *Client) FindChartMatches(ctx context.Context, req *chart.ScanRequest) ([]*scanner.ChartSegment, error) {
	const op = "scanner.Client.FindCandleMatches"

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	resp, err := c.api.FindChartMatches(ctx, mapper.ToChartScanRequest(req))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	res := mapper.FromChartMatchesResponse(resp)
	return res, nil
}

func (c *Client) ComputeCandleStats(ctx context.Context, req *candle.StatsRequest) (*scanner.ScanStats, error) {
	const op = "scanner.Client.ComputeCandleStats"

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	resp, err := c.api.ComputeCandleStats(ctx, mapper.ToCandleStatsRequest(req))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	res := mapper.FromCandleStatsResponse(resp)
	return res, nil
}

func (c *Client) ComputeChartStats(ctx context.Context, req *chart.StatsRequest) (*scanner.ScanStats, error) {
	const op = "scanner.Client.ComputeChartStats"

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.api.ComputeChartStats(ctx, mapper.ToChartStatsRequest(req))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	res := mapper.FromChartStatsResponse(resp)
	return res, nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
