package scanner

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/m1keee3/FinanceAnalyst/pkg/logger/sl"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/domain/models"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/cache"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/mapper"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/scanner/candle"
	candlemodels "github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/scanner/candle/models"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/scanner/chart"
	chartmodels "github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/scanner/chart/models"
	scannerv1 "github.com/m1keee3/FinanceAnalyst/services/scanner/proto-gen/v1"
)

type Cache interface {
	GetScan(ctx context.Context, hash string) ([]models.ChartSegment, error)
	SetScan(ctx context.Context, hash string, segments []models.ChartSegment, ttl time.Duration) error
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

type ScanResult struct {
	matches []models.ChartSegment
	err     error
}

func (s *Service) FindCandleMatches(ctx context.Context, request *scannerv1.CandleScanRequest) (*scannerv1.ScanResponse, error) {
	const op = "ScannerService.FindCandleMatches"

	log := s.log.With(slog.String("op", op))
	log.Info("find candle matches request")

	query := candlemodels.NewScanQuery(request)
	hash := query.Hash()

	cached, err := s.cache.GetScan(ctx, hash)
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			log.Info("no cached matches found")
		} else {
			log.Error("failed to get cached matches", sl.Err(err))
		}
	} else if cached != nil {
		log.Info("found cached matches")
		return matchesToScanResponse(cached), nil
	}

	resCh := make(chan ScanResult, 1)

	go func() {
		matches, err := s.candleScanner.Scan(candlemodels.NewScanQuery(request))
		resCh <- ScanResult{matches, err}
	}()

	select {
	case <-ctx.Done():
		log.Error("context canceled", sl.Err(ctx.Err()))
		return nil, fmt.Errorf("%s: %w", op, ctx.Err())

	case res := <-resCh:
		if res.err != nil {
			log.Error("failed to scan", sl.Err(res.err))
			return nil, fmt.Errorf("%s: %w", op, res.err)
		}

		go func() {
			if err := s.cache.SetScan(ctx, hash, res.matches, s.ttl); err != nil {
				log.Warn("failed to cache matches", sl.Err(err))
			}
		}()

		return matchesToScanResponse(res.matches), nil
	}
}

func (s *Service) FindChartMatches(ctx context.Context, request *scannerv1.ChartScanRequest) (*scannerv1.ScanResponse, error) {
	const op = "ScannerService.FindChartMatches"

	log := s.log.With(slog.String("op", op))
	log.Info("find chart matches request")

	query := chartmodels.NewScanQuery(request)
	hash := query.Hash()

	cached, err := s.cache.GetScan(ctx, hash)
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			log.Info("no cached matches found")
		} else {
			log.Warn("failed to get cached matches", sl.Err(err))
		}
	} else if cached != nil {
		log.Info("found cached matches")
		return matchesToScanResponse(cached), nil
	}

	resCh := make(chan ScanResult, 1)

	go func() {
		matches, err := s.chartScanner.Scan(chartmodels.NewScanQuery(request))
		resCh <- ScanResult{matches, err}
	}()

	select {
	case <-ctx.Done():
		log.Error("context canceled", sl.Err(ctx.Err()))
		return nil, fmt.Errorf("%s: %w", op, ctx.Err())

	case res := <-resCh:
		if res.err != nil {
			log.Error("failed to scan", sl.Err(res.err))
			return nil, fmt.Errorf("%s: %w", op, res.err)
		}

		go func() {
			if err := s.cache.SetScan(ctx, hash, res.matches, s.ttl); err != nil {
				log.Warn("failed to cache matches", sl.Err(err))
			}
		}()

		return matchesToScanResponse(res.matches), nil
	}
}

func (s *Service) ComputeCandleStats(ctx context.Context, request *scannerv1.ComputeStatsCandleRequest) (*scannerv1.ComputeStatsResponse, error) {
	const op = "ScannerService.ComputeCandleStats"

	log := s.log.With(slog.String("op", op))
	log.Info("compute candle stats request")

	query := candlemodels.NewScanQuery(request.GetScan())
	hash := query.Hash()

	cached, err := s.cache.GetScan(ctx, hash)
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			log.Info("no cached matches found")
		} else {
			log.Warn("failed to get cached matches", sl.Err(err))
		}
	} else if cached != nil {
		log.Info("found cached matches")
		stats, err := s.statsComputer.ComputeStats(cached, int(request.GetDaysToWatch()))
		if err != nil {
			log.Error("failed to compute candle stats", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, sl.Err(err))
		}

		return scanStatsToComputeStatsResponse(stats), nil
	}

	resCh := make(chan ScanResult, 1)

	go func() {
		matches, err := s.candleScanner.Scan(candlemodels.NewScanQuery(request.GetScan()))
		resCh <- ScanResult{matches, err}
	}()

	select {
	case <-ctx.Done():
		log.Error("context canceled", sl.Err(ctx.Err()))
		return nil, fmt.Errorf("%s: %w", op, ctx.Err())
	case res := <-resCh:
		if res.err != nil {
			log.Error("failed to compute candle stats", sl.Err(res.err))
			return nil, fmt.Errorf("%s: %w", op, sl.Err(res.err))
		}

		go func() {
			if err := s.cache.SetScan(ctx, hash, res.matches, s.ttl); err != nil {
				log.Warn("failed to cache matches", sl.Err(err))
			}
		}()

		stats, err := s.statsComputer.ComputeStats(res.matches, int(request.GetDaysToWatch()))
		if err != nil {
			log.Error("failed to compute candle stats", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, sl.Err(err))
		}

		return scanStatsToComputeStatsResponse(stats), nil
	}
}

func (s *Service) ComputeChartStats(ctx context.Context, request *scannerv1.ComputeStatsChartRequest) (*scannerv1.ComputeStatsResponse, error) {
	const op = "ScannerService.ComputeChartStats"

	log := s.log.With(slog.String("op", op))
	log.Info("compute chart stats request")

	query := chartmodels.NewScanQuery(request.GetScan())
	hash := query.Hash()

	cached, err := s.cache.GetScan(ctx, hash)
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			log.Info("no cached matches found")
		} else {
			log.Warn("failed to get cached matches", sl.Err(err))
		}
	} else if cached != nil {
		log.Info("found cached matches")
		stats, err := s.statsComputer.ComputeStats(cached, int(request.GetDaysToWatch()))
		if err != nil {
			log.Error("failed to compute chart stats", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, sl.Err(err))
		}

		return scanStatsToComputeStatsResponse(stats), nil
	}

	resCh := make(chan ScanResult, 1)

	go func() {
		matches, err := s.chartScanner.Scan(chartmodels.NewScanQuery(request.GetScan()))
		resCh <- ScanResult{matches, err}
	}()

	select {
	case <-ctx.Done():
		log.Error("context canceled", sl.Err(ctx.Err()))
		return nil, fmt.Errorf("%s: %w", op, ctx.Err())
	case res := <-resCh:
		if res.err != nil {
			log.Error("failed to compute chart stats", sl.Err(res.err))
			return nil, fmt.Errorf("%s: %w", op, sl.Err(res.err))
		}

		go func() {
			if err := s.cache.SetScan(ctx, hash, res.matches, s.ttl); err != nil {
				log.Warn("failed to cache matches", sl.Err(err))
			}
		}()

		stats, err := s.statsComputer.ComputeStats(res.matches, int(request.GetDaysToWatch()))
		if err != nil {
			log.Error("failed to compute chart stats", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, sl.Err(err))
		}

		return scanStatsToComputeStatsResponse(stats), nil
	}
}

func matchesToScanResponse(matches []models.ChartSegment) *scannerv1.ScanResponse {
	protoMatches := make([]*scannerv1.ChartSegment, len(matches))

	for i, m := range matches {
		protoMatches[i] = mapper.ToProtoChartSegment(m)
	}

	return &scannerv1.ScanResponse{
		Matches: protoMatches,
	}
}

func scanStatsToComputeStatsResponse(stats *models.ScanStats) *scannerv1.ComputeStatsResponse {
	return &scannerv1.ComputeStatsResponse{
		Stats: &scannerv1.ScanStats{
			TotalMatches: int32(stats.TotalMatches),
			PriceChange:  stats.PriceChange,
			Probability:  stats.Probability,
		},
	}
}
