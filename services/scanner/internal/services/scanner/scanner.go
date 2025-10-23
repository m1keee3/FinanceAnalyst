package scanner

import (
	"context"
	"log/slog"
	"time"

	"github.com/m1keee3/FinanceAnalyst/pkg/logger/sl"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/domain/models"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/cache"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/mapper"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/scanner/candle"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/scanner/chart"
	scannerv1 "github.com/m1keee3/FinanceAnalyst/services/scanner/proto-gen/v1"
)

type Cache interface {
	GetScan(ctx context.Context, hash string) ([]models.ChartSegment, error)
	SetScan(ctx context.Context, hash string, segments []models.ChartSegment, ttl time.Duration) error

	GetStats(ctx context.Context, hash string) (*models.ScanStats, error)
	SetStats(ctx context.Context, hash string, stats *models.ScanStats, ttl time.Duration) error
}

type StatsComputer interface {
	ComputeStats(matches []models.ChartSegment, daysToWatch int) (*models.ScanStats, error)
}

type Service struct {
	log           *slog.Logger
	candleScanner *candle.Scanner
	chartScanner  *chart.Scanner
	statsComputer StatsComputer
	cache         Cache
	ttl           time.Duration
}

func NewService(
	log *slog.Logger,
	candleScanner *candle.Scanner,
	chartScanner *chart.Scanner,
	statsComputer StatsComputer,
	cache Cache,
	ttl time.Duration,
) *Service {
	return &Service{
		log:           log,
		candleScanner: candleScanner,
		chartScanner:  chartScanner,
		statsComputer: statsComputer,
		cache:         cache,
		ttl:           ttl,
	}
}

func (s *Service) FindCandleMatches(ctx context.Context, request *scannerv1.CandleScanRequest) (*scannerv1.ScanResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Service) FindChartMatches(ctx context.Context, request *scannerv1.ChartScanRequest) (*scannerv1.ScanResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Service) ComputeCandleStats(ctx context.Context, request *scannerv1.ComputeStatsCandleRequest) (*scannerv1.ComputeStatsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Service) ComputeChartStats(ctx context.Context, request *scannerv1.ComputeStatsChartRequest) (*scannerv1.ComputeStatsResponse, error) {
	//TODO implement me
	panic("implement me")
}
