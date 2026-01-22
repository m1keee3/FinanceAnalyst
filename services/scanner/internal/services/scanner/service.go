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

const (
	dayDuration      = 24 * time.Hour
	halfHourDuration = 30 * time.Minute
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
}

func NewService(
	log *slog.Logger,
	candleScanner *candle.Scanner,
	chartScanner *chart.Scanner,
	statsComputer StatsComputer,
	cache Cache,
) *Service {
	return &Service{
		log:           log,
		candleScanner: candleScanner,
		chartScanner:  chartScanner,
		statsComputer: statsComputer,
		cache:         cache,
	}
}

type ScanResult struct {
	matches []models.ChartSegment
	err     error
}

func (s *Service) FindCandleMatches(ctx context.Context, query *candlemodels.ScanQuery) ([]models.ChartSegment, error) {
	const op = "scanner.Service.FindCandleMatches"

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
			if err := s.cache.SetScan(ctx, hash, res.matches, s.candleTTL(query)); err != nil {
				log.Warn("failed to cache matches", sl.Err(err))
			} else {
				log.Info("cached matches")
			}
		}()

		return res.matches, nil
	}
}

func (s *Service) FindChartMatches(ctx context.Context, query *chartmodels.ScanQuery) ([]models.ChartSegment, error) {
	const op = "scanner.Service.FindChartMatches"

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
			if err := s.cache.SetScan(ctx, hash, res.matches, s.chartTTL(query)); err != nil {
				log.Warn("failed to cache matches", sl.Err(err))
			} else {
				log.Info("cached matches")
			}
		}()

		return res.matches, nil
	}
}

func (s *Service) ComputeCandleStats(ctx context.Context, query *candlemodels.StatsQuery) (*models.ScanStats, error) {
	const op = "scanner.Service.ComputeCandleStats"

	log := s.log.With(slog.String("op", op))
	log.Info("compute candle stats request")

	cachedStats, err := s.cache.GetStats(ctx, query.Hash())
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			log.Info("no cached stats found")
		} else {
			log.Warn("failed to get cached matches", sl.Err(err))
		}
	} else if cachedStats != nil {
		log.Info("found cached stats")
		return cachedStats, nil
	}

	cachedScan, err := s.cache.GetScan(ctx, query.ScanQuery.Hash())
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			log.Info("no cached matches found")
		} else {
			log.Warn("failed to get cached matches", sl.Err(err))
		}
	} else if cachedScan != nil {
		log.Info("found cached matches")
		stats, err := s.statsComputer.ComputeStats(cachedScan, query.DaysToWatch)
		if err != nil {
			log.Error("failed to compute chart stats", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, sl.Err(err))
		}

		return stats, nil
	}

	resCh := make(chan ScanResult, 1)

	go func() {
		matches, err := s.candleScanner.Scan(query.ScanQuery)
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
			if err := s.cache.SetScan(ctx, query.ScanQuery.Hash(), res.matches, s.candleTTL(query.ScanQuery)); err != nil {
				log.Warn("failed to cache matches", sl.Err(err))
			} else {
				log.Info("cached matches")
			}
		}()

		stats, err := s.statsComputer.ComputeStats(res.matches, query.DaysToWatch)
		if err != nil {
			log.Error("failed to compute candle stats", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, sl.Err(err))
		}

		go func() {
			if err := s.cache.SetStats(ctx, query.Hash(), stats, s.candleTTL(query.ScanQuery)); err != nil {
				log.Warn("failed to cache stats", sl.Err(err))
			} else {
				log.Info("cached stats")
			}
		}()

		return stats, nil
	}
}

func (s *Service) ComputeChartStats(ctx context.Context, query *chartmodels.StatsQuery) (*models.ScanStats, error) {
	const op = "scanner.Service.ComputeChartStats"

	log := s.log.With(slog.String("op", op))
	log.Info("compute chart stats request")

	hash := query.Hash()

	cachedStats, err := s.cache.GetStats(ctx, query.Hash())
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			log.Info("no cached stats found")
		} else {
			log.Warn("failed to get cached matches", sl.Err(err))
		}
	} else if cachedStats != nil {
		log.Info("found cached stats")
		return cachedStats, nil
	}

	cachedScan, err := s.cache.GetScan(ctx, query.ScanQuery.Hash())
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			log.Info("no cached matches found")
		} else {
			log.Warn("failed to get cached matches", sl.Err(err))
		}
	} else if cachedScan != nil {
		log.Info("found cached matches")
		stats, err := s.statsComputer.ComputeStats(cachedScan, query.DaysToWatch)
		if err != nil {
			log.Error("failed to compute chart stats", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, sl.Err(err))
		}

		return stats, nil
	}

	resCh := make(chan ScanResult, 1)

	go func() {
		matches, err := s.chartScanner.Scan(query.ScanQuery)
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
			if err := s.cache.SetScan(ctx, hash, res.matches, s.chartTTL(query.ScanQuery)); err != nil {
				log.Warn("failed to cache matches", sl.Err(err))
			} else {
				log.Info("cached matches")
			}
		}()

		stats, err := s.statsComputer.ComputeStats(res.matches, query.DaysToWatch)
		if err != nil {
			log.Error("failed to compute chart stats", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, sl.Err(err))
		}

		go func() {
			if err := s.cache.SetStats(ctx, query.Hash(), stats, s.chartTTL(query.ScanQuery)); err != nil {
				log.Warn("failed to cache stats", sl.Err(err))
			} else {
				log.Info("cached stats")
			}
		}()

		return stats, nil
	}
}

func (s *Service) candleTTL(query *candlemodels.ScanQuery) time.Duration {
	ttl := dayDuration
	if sameDay(query.SearchTo, time.Now()) || sameDay(query.Segment.To, time.Now()) {
		ttl = halfHourDuration
	}

	return ttl
}

func (s *Service) chartTTL(query *chartmodels.ScanQuery) time.Duration {
	ttl := dayDuration
	if sameDay(query.SearchTo, time.Now()) || sameDay(query.Segment.To, time.Now()) {
		ttl = halfHourDuration
	}

	return ttl
}

func sameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
