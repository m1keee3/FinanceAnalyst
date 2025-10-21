package scannergrpc

import (
	"context"

	scannerv1 "github.com/m1keee3/FinanceAnalyst/services/scanner/proto-gen/v1"
	"google.golang.org/grpc"
)

type serverAPI struct {
	scannerv1.UnimplementedScannerServiceServer
}

func Register(grpcServer *grpc.Server) {
	scannerv1.RegisterScannerServiceServer(grpcServer, &serverAPI{})
}

/*type Scanner interface {
	// Поиск свечных паттернов
	FindCandleMatches(context.Context, *scannerv1.CandleScanRequest) (*scannerv1.ScanResponse, error)
	// Поиск графических паттернов
	FindChartMatches(context.Context, *scannerv1.ChartScanRequest) (*scannerv1.ScanResponse, error)
	// Вычисление статистики для свечных паттернов
	ComputeCandleStats(context.Context, *scannerv1.ComputeStatsCandleRequest) (*scannerv1.ComputeStatsResponse, error)
	// Вычисление статистики для графических паттернов
	ComputeChartStats(context.Context, *scannerv1.ComputeStatsChartRequest) (*scannerv1.ComputeStatsResponse, error)
}*/

func (s *serverAPI) FindCandleMatches(ctx context.Context, request *scannerv1.CandleScanRequest) (*scannerv1.ScanResponse, error) {
	panic("implement me")
}

func (s *serverAPI) FindChartMatches(ctx context.Context, request *scannerv1.ChartScanRequest) (*scannerv1.ScanResponse, error) {
	panic("implement me")
}

func (s *serverAPI) ComputeCandleStats(ctx context.Context, request *scannerv1.ComputeStatsCandleRequest) (*scannerv1.ComputeStatsResponse, error) {
	panic("implement me")
}

func (s *serverAPI) ComputeChartStats(ctx context.Context, request *scannerv1.ComputeStatsChartRequest) (*scannerv1.ComputeStatsResponse, error) {
	panic("implement me")
}
