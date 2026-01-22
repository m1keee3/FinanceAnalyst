package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/m1keee3/FinanceAnalyst/services/watcher/domain/models"
	"github.com/m1keee3/FinanceAnalyst/services/watcher/internal/storage"
	"time"
)

type Storage struct {
	db *sql.DB
}

func New(dsn string) (*Storage, error) {
	const op = "storage.Postgres.New"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) GetDailyStats(ctx context.Context) ([]models.SegmentStats, error) {
	const op = "storage.Postgres.GetDailyStats"

	query := `
		SELECT
			ticker,
			segment_from,
			segment_to,
			total_matches,
			price_change,
			probability,
			pattern_type
		FROM segment_stats
		WHERE stat_date = $1
	`

	statDate := time.Now().UTC().Truncate(24 * time.Hour)

	rows, err := s.db.QueryContext(ctx, query, statDate)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var result []models.SegmentStats

	for rows.Next() {
		var (
			ticker      string
			from, to    time.Time
			total       int
			priceChange float64
			probability float64
			patternType models.PatternType
		)

		if err := rows.Scan(
			&ticker,
			&from,
			&to,
			&total,
			&priceChange,
			&probability,
			&patternType,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		result = append(result, models.SegmentStats{
			Segment: &models.ChartSegment{
				Ticker: ticker,
				From:   from,
				To:     to,
			},
			TotalMatches: total,
			PriceChange:  priceChange,
			Probability:  probability,
			PatternType:  patternType,
		})
	}

	if len(result) == 0 {
		return nil, storage.ErrNotFound
	}

	return result, nil
}

func (s *Storage) SetDailySegmentStats(ctx context.Context, segStats models.SegmentStats) error {
	const op = "storage.Postgres.SetDailySegmentStats"

	query := `
		INSERT INTO segment_stats (
			stat_date,
			ticker,
			pattern_type,
			segment_from,
			segment_to,
			total_matches,
			price_change,
			probability
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (stat_date, ticker, pattern_type)
		DO UPDATE SET
			segment_from   = EXCLUDED.segment_from,
			segment_to     = EXCLUDED.segment_to,
			total_matches  = EXCLUDED.total_matches,
			price_change   = EXCLUDED.price_change,
			probability    = EXCLUDED.probability,
			updated_at     = now()
	`

	statDate := time.Now().UTC().Truncate(24 * time.Hour)

	_, err := s.db.ExecContext(
		ctx,
		query,
		statDate,
		segStats.Segment.Ticker,
		segStats.PatternType,
		segStats.Segment.From,
		segStats.Segment.To,
		segStats.TotalMatches,
		segStats.PriceChange,
		segStats.Probability,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) Close() error {
	const op = "storage.Postgres.Close"

	if err := s.db.Close(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
