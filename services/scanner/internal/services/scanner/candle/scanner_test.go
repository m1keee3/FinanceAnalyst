package candle

import (
	"log/slog"
	"testing"
	"time"

	"github.com/m1keee3/FinanceAnalyst/services/scanner/domain/models"
	candlemodels "github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/models/candle"
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

func candleAt(ts int64, open, close, high, low float64) models.Candle {
	return models.Candle{
		Date:  time.Unix(ts, 0),
		Open:  open,
		Close: close,
		High:  high,
		Low:   low,
	}
}

func upCandles(start int64, n int) []models.Candle {
	res := make([]models.Candle, 0, n)
	for i := 0; i < n; i++ {
		res = append(res, candleAt(
			start+int64(i*60),
			1+float64(i),
			2+float64(i),
			2.2+float64(i),
			0.8+float64(i),
		))
	}
	return res
}

func downCandles(start int64, n int) []models.Candle {
	res := make([]models.Candle, 0, n)
	for i := 0; i < n; i++ {
		res = append(res, candleAt(
			start+int64(i*60),
			2+float64(i),
			1+float64(i),
			2.2+float64(i),
			0.8+float64(i),
		))
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
		query   *candlemodels.ScanQuery
	}{
		{
			name:    "nil scanner",
			scanner: nil,
			query:   &candlemodels.ScanQuery{},
		},
		{
			name:    "nil query",
			scanner: NewScanner(logger, &mockFetcher{}),
			query:   nil,
		},
		{
			name:    "empty tickers",
			scanner: NewScanner(logger, &mockFetcher{data: map[string][]models.Candle{"A": upCandles(0, 2)}}),
			query: &candlemodels.ScanQuery{
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

func TestScanner_Scan_FindsSingleMatch(t *testing.T) {
	logger := slog.Default()

	pattern := upCandles(0, 3)
	search := upCandles(0, 5)

	fetcher := &mockFetcher{
		data: map[string][]models.Candle{
			"PATTERN": pattern,
			"AAPL":    search,
		},
	}

	scanner := NewScanner(logger, fetcher)

	query := &candlemodels.ScanQuery{
		Segment: models.ChartSegment{
			Ticker: "PATTERN",
			From:   pattern[0].Date,
			To:     pattern[len(pattern)-1].Date,
		},
		Tickers:    []string{"AAPL"},
		SearchFrom: search[0].Date,
		SearchTo:   search[len(search)-1].Date,
	}

	matches, err := scanner.Scan(query)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(matches) != 3 {
		t.Fatalf("expected 3 match, got %d", len(matches))
	}
}

func TestScanner_Scan_TailMismatch(t *testing.T) {
	logger := slog.Default()

	pattern := upCandles(0, 3)
	search := downCandles(0, 5)

	fetcher := &mockFetcher{
		data: map[string][]models.Candle{
			"PATTERN": pattern,
			"TSLA":    search,
		},
	}

	scanner := NewScanner(logger, fetcher)

	query := &candlemodels.ScanQuery{
		Segment: models.ChartSegment{
			Ticker: "PATTERN",
			From:   pattern[0].Date,
			To:     pattern[len(pattern)-1].Date,
		},
		Options: candlemodels.ScanOptions{
			TailLen: 1,
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

func TestScanner_Scan_ToleranceReject(t *testing.T) {
	logger := slog.Default()

	pattern := upCandles(0, 2)
	search := []models.Candle{
		candleAt(0, 10, 20, 21, 9),
		candleAt(60, 30, 10, 31, 9), // направление отличается
	}

	fetcher := &mockFetcher{
		data: map[string][]models.Candle{
			"PATTERN": pattern,
			"MSFT":    search,
		},
	}

	scanner := NewScanner(logger, fetcher)

	query := &candlemodels.ScanQuery{
		Segment: models.ChartSegment{
			Ticker: "PATTERN",
			From:   pattern[0].Date,
			To:     pattern[len(pattern)-1].Date,
		},
		Options: candlemodels.ScanOptions{
			BodyTolerance:   0.01,
			ShadowTolerance: 0.01,
		},
		Tickers:    []string{"MSFT"},
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

func TestScanner_Scan_OverlapIgnored(t *testing.T) {
	logger := slog.Default()

	pattern := upCandles(0, 2)

	fetcher := &mockFetcher{
		data: map[string][]models.Candle{
			"AAPL": pattern,
		},
	}

	scanner := NewScanner(logger, fetcher)

	query := &candlemodels.ScanQuery{
		Segment: models.ChartSegment{
			Ticker: "AAPL",
			From:   pattern[0].Date,
			To:     pattern[len(pattern)-1].Date,
		},
		Tickers:    []string{"AAPL"},
		SearchFrom: pattern[0].Date,
		SearchTo:   pattern[len(pattern)-1].Date,
	}

	matches, err := scanner.Scan(query)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(matches) != 0 {
		t.Fatalf("expected 0 matches due to overlap, got %d", len(matches))
	}
}
