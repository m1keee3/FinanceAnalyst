package scannergrpc

import (
	"context"

	"github.com/m1keee3/FinanceAnalyst/services/scanner/domain/models"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/grpc/mapper"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/models/candle"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/models/chart"
	scannerv1 "github.com/m1keee3/FinanceAnalyst/services/scanner/proto-gen/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ScannerService interface {
	FindCandleMatches(ctx context.Context, query *candle.ScanQuery) ([]models.ChartSegment, error)

	FindChartMatches(ctx context.Context, query *chart.ScanQuery) ([]models.ChartSegment, error)

	ComputeCandleStats(ctx context.Context, query *candle.ScanQuery, daysToWatch int) (*models.ScanStats, error)

	ComputeChartStats(ctx context.Context, query *chart.ScanQuery, daysToWatch int) (*models.ScanStats, error)
}

type serverAPI struct {
	scannerv1.UnimplementedScannerServer
	scannerService ScannerService
}

func Register(grpcServer *grpc.Server, scannerService ScannerService) {
	scannerv1.RegisterScannerServer(grpcServer, &serverAPI{scannerService: scannerService})
}

func (s *serverAPI) FindCandleMatches(ctx context.Context, request *scannerv1.CandleScanRequest) (*scannerv1.CandleMatchesResponse, error) {
	if request.Segment == nil {
		return nil, status.Errorf(codes.InvalidArgument, "segment is required")
	}

	if request.SearchFrom == nil {
		return nil, status.Errorf(codes.InvalidArgument, "search from is required")
	}

	if request.SearchTo == nil {
		return nil, status.Errorf(codes.InvalidArgument, "search to is required")
	}

	if len(request.Tickers) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one ticker is required")
	}

	query := mapper.CandleScanRequestToScanQuery(request)

	matches, err := s.scannerService.FindCandleMatches(ctx, query)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to find candle matches")
	}

	return mapper.MatchesToCandleMatchResponse(matches), nil
}

func (s *serverAPI) FindChartMatches(ctx context.Context, request *scannerv1.ChartScanRequest) (*scannerv1.ChartMatchesResponse, error) {
	if request.Segment == nil {
		return nil, status.Errorf(codes.InvalidArgument, "segment is required")
	}

	if request.SearchFrom == nil {
		return nil, status.Errorf(codes.InvalidArgument, "search from is required")
	}

	if request.SearchTo == nil {
		return nil, status.Errorf(codes.InvalidArgument, "search to is required")
	}

	if len(request.Tickers) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one ticker is required")
	}

	query := mapper.ChartScanRequestToScanQuery(request)

	matches, err := s.scannerService.FindChartMatches(ctx, query)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to find chart matches")
	}

	return mapper.MatchesToChartMatchResponse(matches), nil
}

func (s *serverAPI) ComputeCandleStats(ctx context.Context, request *scannerv1.CandleStatsRequest) (*scannerv1.CandleStatsResponse, error) {
	if request.Scan.Segment == nil {
		return nil, status.Errorf(codes.InvalidArgument, "segment is required")
	}

	if request.Scan.SearchFrom == nil {
		return nil, status.Errorf(codes.InvalidArgument, "search from is required")
	}

	if request.Scan.SearchTo == nil {
		return nil, status.Errorf(codes.InvalidArgument, "search to is required")
	}

	if len(request.Scan.Tickers) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one ticker is required")
	}

	daysToWatch := int(request.DaysToWatch)
	if daysToWatch == 0 {
		daysToWatch = 1
	}

	query := mapper.CandleScanRequestToScanQuery(request.Scan)

	scanStats, err := s.scannerService.ComputeCandleStats(ctx, query, daysToWatch)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to compute candle stats")
	}

	return mapper.ScanStatsToCandleStatsResponse(scanStats), nil
}

func (s *serverAPI) ComputeChartStats(ctx context.Context, request *scannerv1.ChartStatsRequest) (*scannerv1.ChartStatsResponse, error) {
	if request.Scan.Segment == nil {
		return nil, status.Errorf(codes.InvalidArgument, "segment is required")
	}

	if request.Scan.SearchFrom == nil {
		return nil, status.Errorf(codes.InvalidArgument, "search from is required")
	}

	if request.Scan.SearchTo == nil {
		return nil, status.Errorf(codes.InvalidArgument, "search to is required")
	}

	if len(request.Scan.Tickers) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one ticker is required")
	}

	daysToWatch := int(request.DaysToWatch)
	if daysToWatch == 0 {
		daysToWatch = 1
	}

	query := mapper.ChartScanRequestToScanQuery(request.Scan)

	scanStats, err := s.scannerService.ComputeChartStats(ctx, query, daysToWatch)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to compute chart stats")
	}

	return mapper.ScanStatsToChartStatsResponse(scanStats), nil
}
