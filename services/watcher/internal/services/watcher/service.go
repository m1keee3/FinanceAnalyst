package watcher

import (
	"context"
	"errors"
	"github.com/m1keee3/FinanceAnalyst/pkg/logger/sl"
	"github.com/m1keee3/FinanceAnalyst/services/watcher/domain/models"
	"github.com/m1keee3/FinanceAnalyst/services/watcher/internal/cache"
	"github.com/m1keee3/FinanceAnalyst/services/watcher/internal/config"
	"github.com/m1keee3/FinanceAnalyst/services/watcher/internal/services/watcher/models/candle"
	"github.com/m1keee3/FinanceAnalyst/services/watcher/internal/services/watcher/models/chart"
	"log/slog"
	"sync"
	"time"
)

type Publisher interface {
	PublishCandleStats(segStats *models.SegmentStats) error
	PublishChartStats(segStats *models.SegmentStats) error
}

type Cache interface {
	GetStats(ctx context.Context) ([]models.SegmentStats, error)
	SetSegmentStats(ctx context.Context, segStats models.SegmentStats) error
	Clear() error
}

type Scanner interface {
	ComputeCandleStats(ctx context.Context, query *candle.ScanQuery) (*models.SegmentStats, error)
	ComputeChartStats(ctx context.Context, query *chart.ScanQuery) (*models.SegmentStats, error)
}

type Service struct {
	log       *slog.Logger
	cfg       *config.AppConfig
	publisher Publisher
	cache     Cache
	scanner   Scanner
}

func New(
	log *slog.Logger,
	cfg *config.AppConfig,
	publisher Publisher,
	cache Cache,
	scanner Scanner,
) *Service {
	return &Service{
		log:       log,
		cfg:       cfg,
		publisher: publisher,
		cache:     cache,
		scanner:   scanner,
	}
}

func (s *Service) GetStats(ctx context.Context) []models.SegmentStats {
	const op = "watcher.Service.GetStats"

	log := s.log.With(slog.String("op", op))
	log.Info("get stats request")
	stats, err := s.cache.GetStats(ctx)
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			log.Warn("stats not found in cache")
			return nil
		}
		s.log.Error("failed to get stats from cache", sl.Err(err))
		return nil
	}
	return stats
}

func (s *Service) Run(ctx context.Context) {
	const op = "watcher.Service.Run"

	log := s.log.With(slog.String("op", op))
	log.Info("watcher service run started")

	if err := s.cache.Clear(); err != nil {
		log.Error("failed to clear cache", sl.Err(err))
	}

	done := make(chan struct{})
	var wg sync.WaitGroup

	for _, ticker := range s.cfg.Tickers {
		wg.Add(1)
		go func(ticker string) {
			defer wg.Done()
			s.runCandleStatsForTicker(ctx, ticker)
			s.runChartStatsForTicker(ctx, ticker)
		}(ticker)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	for {
		select {
		case <-done:
			log.Info("watcher service run completed")
			return
		case <-ctx.Done():
			log.Warn("watcher service run canceled, waiting for jobs to finish")
			wg.Wait()
			return
		}
	}
}

func (s *Service) runCandleStatsForTicker(ctx context.Context, ticker string) {
	const op = "watcher.Service.runCandleStatsForTicker"

	log := s.log.With(slog.String("op", op), slog.String("ticker", ticker))
	log.Info("running candle stats for ticker")

	var statsToHandle *models.SegmentStats

	for tailLen := s.cfg.CandleOptions.MinTailLen; tailLen <= s.cfg.CandleOptions.MaxTailLen; tailLen++ {

		var res []models.SegmentStats

		for segLen := s.cfg.CandleMinLen; segLen <= s.cfg.CandleMaxLen; segLen++ {

			stats, err := s.scanner.ComputeCandleStats(ctx,
				&candle.ScanQuery{
					Segment: &models.ChartSegment{
						Ticker: ticker,
						From:   time.Now().AddDate(0, 0, -segLen),
						To:     time.Now(),
					},
					Options: &candle.ScanOptions{
						TailLen:         tailLen,
						ShadowTolerance: s.cfg.CandleOptions.ShadowTolerance,
						BodyTolerance:   s.cfg.CandleOptions.BodyTolerance,
					},
					SearchFrom: s.cfg.SearchFrom,
					SearchTo:   s.cfg.SearchTo,
					Tickers:    []string{ticker},
				})
			if err != nil {
				log.Error("failed to compute candle stats", sl.Err(err))
			}

			if stats.TotalMatches >= s.cfg.CandleOptions.MinMatches {
				res = append(res, *stats)
			}
		}

		// Among stats with TotalMatches >= MinMatches, prefer with the highest TailLen
		if len(res) > 0 {
			maxMatches := 0
			for i := range res {
				if res[i].TotalMatches >= maxMatches {
					maxMatches = res[i].TotalMatches
					statsToHandle = &res[i]
				}
			}
		}
	}

	if statsToHandle != nil {
		s.handleCandleStats(ctx, statsToHandle)
	}
}

func (s *Service) runChartStatsForTicker(ctx context.Context, ticker string) {
	const op = "watcher.Service.runChartStatsForTicker"

	log := s.log.With(slog.String("op", op), slog.String("ticker", ticker))
	log.Info("running chart stats for ticker")

	var res []models.SegmentStats

	for segLen := s.cfg.ChartMinLen; segLen <= s.cfg.ChartMaxLen; segLen++ {

		stats, err := s.scanner.ComputeChartStats(ctx,
			&chart.ScanQuery{
				Segment: &models.ChartSegment{
					Ticker: ticker,
					From:   time.Now().AddDate(0, 0, -segLen),
					To:     time.Now(),
				},
				Options: &chart.ScanOptions{
					MinScale:    s.cfg.ChartOptions.MinScale,
					MaxScale:    s.cfg.ChartOptions.MaxScale,
					Tolerance:   s.cfg.ChartOptions.Tolerance,
					DaysToWatch: s.cfg.ChartOptions.DaysToWatch,
				},
				SearchFrom: s.cfg.SearchFrom,
				SearchTo:   s.cfg.SearchTo,
				Tickers:    []string{ticker},
			})

		if err != nil {
			log.Error("failed to compute chart stats", sl.Err(err))
		}

		if stats.TotalMatches >= s.cfg.ChartOptions.MinMatches {
			res = append(res, *stats)
		}
	}

	var statsToHandle *models.SegmentStats

	maxMatches := 0
	for i := range res {
		if res[i].TotalMatches >= maxMatches {
			maxMatches = res[i].TotalMatches
			statsToHandle = &res[i]
		}
	}

	s.handleChartStats(ctx, statsToHandle)
}

func (s *Service) handleCandleStats(ctx context.Context, stats *models.SegmentStats) {
	const op = "watcher.Service.handleCandleStats"

	log := s.log.With(slog.String("op", op))

	if err := s.publisher.PublishCandleStats(stats); err != nil {
		log.Error("failed to publish candle stats", sl.Err(err))
	}

	if err := s.cache.SetSegmentStats(ctx, *stats); err != nil {
		log.Error("failed to cache candle stats", sl.Err(err))
	}
}

func (s *Service) handleChartStats(ctx context.Context, stats *models.SegmentStats) {
	const op = "watcher.Service.handleChartStats"

	log := s.log.With(slog.String("op", op))

	if err := s.publisher.PublishChartStats(stats); err != nil {
		log.Error("failed to publish chart stats", sl.Err(err))
	}

	if err := s.cache.SetSegmentStats(ctx, *stats); err != nil {
		log.Error("failed to cache chart stats", sl.Err(err))
	}
}
