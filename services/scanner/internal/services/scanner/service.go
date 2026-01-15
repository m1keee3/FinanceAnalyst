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
	candlemodels "github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/models/candle"
	chartmodels "github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/models/chart"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/scanner/candle"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/scanner/chart"
)

type Cache interface {
	GetScan(ctx context.Context, hash string) ([]models.ChartSegment, error)
	SetScan(ctx context.Context, hash string, segments []models.ChartSegment, ttl time.Duration) error
}

type TTLResolver interface {
	CandleTTL(searchTo time.Time) time.Duration
	ChartTTL(searchTo time.Time) time.Duration
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
	ttlResolver   TTLResolver
}

func NewService(
	log *slog.Logger,
	candleScanner *candle.Scanner,
	chartScanner *chart.Scanner,
	statsComputer StatsComputer,
	cache Cache,
	ttlResolver TTLResolver,
) *Service {
	return &Service{
		log:           log,
		candleScanner: candleScanner,
		chartScanner:  chartScanner,
		statsComputer: statsComputer,
		cache:         cache,
		ttlResolver:   ttlResolver,
	}
}

type ScanResult struct {
	matches []models.ChartSegment
	err     error
}

func (s *Service) FindCandleMatches(ctx context.Context, query *candlemodels.ScanQuery) ([]models.ChartSegment, error) {
	const op = "ScannerService.FindCandleMatches"

	log := s.log.With(slog.String("op", op))
	log.Info("find candle matches request")

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
		return cached, nil
	}

	resCh := make(chan ScanResult, 1)

	go func() {
		matches, err := s.candleScanner.Scan(query)
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
			if err := s.cache.SetScan(ctx, hash, res.matches, s.ttlResolver.CandleTTL(query.SearchTo)); err != nil {
				log.Warn("failed to cache matches", sl.Err(err))
			}
		}()

		return res.matches, nil
	}
}

func (s *Service) FindChartMatches(ctx context.Context, query *chartmodels.ScanQuery) ([]models.ChartSegment, error) {
	const op = "ScannerService.FindChartMatches"

	log := s.log.With(slog.String("op", op))
	log.Info("find chart matches request")

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
		return cached, nil
	}

	resCh := make(chan ScanResult, 1)

	go func() {
		matches, err := s.chartScanner.Scan(query)
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
			if err := s.cache.SetScan(ctx, hash, res.matches, s.ttlResolver.ChartTTL(query.SearchTo)); err != nil {
				log.Warn("failed to cache matches", sl.Err(err))
			}
		}()

		return res.matches, nil
	}
}

func (s *Service) ComputeCandleStats(ctx context.Context, query *candlemodels.ScanQuery, daysToWatch int) (*models.ScanStats, error) {
	const op = "ScannerService.ComputeCandleStats"

	log := s.log.With(slog.String("op", op))
	log.Info("compute candle stats request")

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
		stats, err := s.statsComputer.ComputeStats(cached, daysToWatch)
		if err != nil {
			log.Error("failed to compute candle stats", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, sl.Err(err))
		}

		return stats, nil
	}

	resCh := make(chan ScanResult, 1)

	go func() {
		matches, err := s.candleScanner.Scan(query)
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
			if err := s.cache.SetScan(ctx, hash, res.matches, s.ttlResolver.CandleTTL(query.SearchTo)); err != nil {
				log.Warn("failed to cache matches", sl.Err(err))
			}
		}()

		stats, err := s.statsComputer.ComputeStats(res.matches, daysToWatch)
		if err != nil {
			log.Error("failed to compute candle stats", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, sl.Err(err))
		}

		return stats, nil
	}
}

func (s *Service) ComputeChartStats(ctx context.Context, query *chartmodels.ScanQuery, daysToWatch int) (*models.ScanStats, error) {
	const op = "ScannerService.ComputeChartStats"

	log := s.log.With(slog.String("op", op))
	log.Info("compute chart stats request")

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
		stats, err := s.statsComputer.ComputeStats(cached, daysToWatch)
		if err != nil {
			log.Error("failed to compute chart stats", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, sl.Err(err))
		}

		return stats, nil
	}

	resCh := make(chan ScanResult, 1)

	go func() {
		matches, err := s.chartScanner.Scan(query)
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
			if err := s.cache.SetScan(ctx, hash, res.matches, s.ttlResolver.ChartTTL(query.SearchTo)); err != nil {
				log.Warn("failed to cache matches", sl.Err(err))
			}
		}()

		stats, err := s.statsComputer.ComputeStats(res.matches, daysToWatch)
		if err != nil {
			log.Error("failed to compute chart stats", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, sl.Err(err))
		}

		return stats, nil
	}
}
