package chart

import (
	"log/slog"
	"testing"
	"time"

	"github.com/m1keee3/FinanceAnalyst/services/scanner/domain/models"
	chartmodels "github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/models/chart"
)

//
// ===== Mock Fetcher =====
//

type mockFetcher struct {
	data map[string][]models.Candle
	err  error
}

func (m *mockFetcher) Fetch(ticker string, from, to time.Time) ([]models.Candle, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data[ticker], nil
}

//
// ===== Helpers =====
//

func candle(ts int64, close float64) models.Candle {
	return models.Candle{
		Date:  time.Unix(ts, 0),
		Open:  close,
		High:  close,
		Low:   close,
		Close: close,
	}
}

func linearCandles(start int64, n int, slope float64) []models.Candle {
	res := make([]models.Candle, n)
	for i := 0; i < n; i++ {
		noise := float64(i%2) * 0.01
		price := float64(i)*slope + noise
		res[i] = candle(start+int64(i*60), price)
	}
	return res
}

//
// ===== Tests =====
//

func TestScanner_Scan_Errors(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name    string
		scanner *Scanner
		query   *chartmodels.ScanQuery
	}{
		{
			name:    "nil scanner",
			scanner: nil,
			query:   &chartmodels.ScanQuery{},
		},
		{
			name:    "nil query",
			scanner: NewScanner(logger, &mockFetcher{}),
			query:   nil,
		},
		{
			name: "empty tickers",
			scanner: NewScanner(logger, &mockFetcher{
				data: map[string][]models.Candle{
					"A": linearCandles(0, 10, 1),
				},
			}),
			query: &chartmodels.ScanQuery{
				Segment: models.ChartSegment{Ticker: "A"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.scanner.Scan(tt.query)
			if err == nil {
				t.Fatalf("expected error")
			}
		})
	}
}

func TestScanner_Scan_ToleranceReject(t *testing.T) {
	logger := slog.Default()

	seed := linearCandles(0, 10, 1)
	search := linearCandles(0, 20, -1)

	fetcher := &mockFetcher{
		data: map[string][]models.Candle{
			"SEED": seed,
			"TSLA": search,
		},
	}

	scanner := NewScanner(logger, fetcher)

	query := &chartmodels.ScanQuery{
		Segment: models.ChartSegment{
			Ticker: "SEED",
			From:   seed[0].Date,
			To:     seed[len(seed)-1].Date,
		},
		Options: chartmodels.ScanOptions{
			Tolerance: 0.05,
		},
		Tickers:    []string{"TSLA"},
		SearchFrom: search[0].Date,
		SearchTo:   search[len(search)-1].Date,
	}

	matches, err := scanner.Scan(query)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(matches) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(matches))
	}
}

func TestScanner_Scan_SeedSegmentRemoved(t *testing.T) {
	logger := slog.Default()

	seed := linearCandles(0, 10, 1)

	fetcher := &mockFetcher{
		data: map[string][]models.Candle{
			"AAPL": seed,
		},
	}

	scanner := NewScanner(logger, fetcher)

	query := &chartmodels.ScanQuery{
		Segment: models.ChartSegment{
			Ticker: "AAPL",
			From:   seed[0].Date,
			To:     seed[len(seed)-1].Date,
		},
		Tickers:    []string{"AAPL"},
		SearchFrom: seed[0].Date,
		SearchTo:   seed[len(seed)-1].Date,
	}

	matches, err := scanner.Scan(query)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(matches) != 0 {
		t.Fatalf("expected seed segment to be removed")
	}
}

func TestScanner_Scan_OverlapResolvedByDistance(t *testing.T) {
	logger := slog.Default()

	seed := linearCandles(0, 10, 1)

	search := append(
		linearCandles(0, 10, 1),
		linearCandles(600, 10, 1.1)...,
	)

	fetcher := &mockFetcher{
		data: map[string][]models.Candle{
			"SEED": seed,
			"NVDA": search,
		},
	}

	scanner := NewScanner(logger, fetcher)

	query := &chartmodels.ScanQuery{
		Segment: models.ChartSegment{
			Ticker: "SEED",
			From:   seed[0].Date,
			To:     seed[len(seed)-1].Date,
		},
		Tickers:    []string{"NVDA"},
		SearchFrom: search[0].Date,
		SearchTo:   search[len(search)-1].Date,
	}

	matches, err := scanner.Scan(query)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(matches) > 1 {
		t.Fatalf("expected overlap resolution, got %d matches", len(matches))
	}
}
