package chart

import (
	"testing"
	"time"

	"github.com/m1keee3/FinanceAnalyst/services/scanner/domain/models"
	chartmodels "github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/scanner/chart/models"
)

// MockFetcher для тестирования
type MockFetcher struct {
	data map[string][]models.Candle
}

func NewMockFetcher() *MockFetcher {
	return &MockFetcher{
		data: make(map[string][]models.Candle),
	}
}

func (m *MockFetcher) AddData(ticker string, candles []models.Candle) {
	m.data[ticker] = candles
}

func (m *MockFetcher) Fetch(ticker string, from, to time.Time) ([]models.Candle, error) {
	return m.data[ticker], nil
}

// createTestCandles создает тестовые свечи с заданным паттерном
func createTestCandles(count int, basePrice float64, pattern string) []models.Candle {
	candles := make([]models.Candle, count)
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < count; i++ {
		price := basePrice
		switch pattern {
		case "up":
			price = basePrice + float64(i)*0.1
		case "down":
			price = basePrice - float64(i)*0.1
		case "volatile":
			if i%2 == 0 {
				price = basePrice + float64(i)*0.2
			} else {
				price = basePrice - float64(i)*0.1
			}
		case "flat":
			price = basePrice
		}

		candles[i] = models.Candle{
			Date:  baseTime.Add(time.Duration(i*24) * time.Hour),
			Open:  price,
			High:  price + 0.5,
			Low:   price - 0.5,
			Close: price,
		}
	}
	return candles
}

// Граничные случаи

func TestScan_NilScanner(t *testing.T) {
	var scanner *Scanner

	query := &chartmodels.ScanQuery{
		Segment: models.ChartSegment{
			Candles: createTestCandles(10, 100.0, "up"),
		},
		Tickers:    []string{"SBER"},
		SearchFrom: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		SearchTo:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	results, err := scanner.Scan(query)

	if err != nil {
		t.Errorf("Scan() error = %v, want nil", err)
	}
	if results != nil {
		t.Errorf("Scan() returned %v, want nil", results)
	}
}

func TestScan_NilFetcher(t *testing.T) {
	scanner := &Scanner{fetcher: nil}

	query := &chartmodels.ScanQuery{
		Segment: models.ChartSegment{
			Candles: createTestCandles(10, 100.0, "up"),
		},
		Tickers:    []string{"SBER"},
		SearchFrom: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		SearchTo:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	results, err := scanner.Scan(query)

	if err != nil {
		t.Errorf("Scan() error = %v, want nil", err)
	}
	if results != nil {
		t.Errorf("Scan() returned %v, want nil", results)
	}
}

func TestScan_EmptySegment(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	query := &chartmodels.ScanQuery{
		Segment: models.ChartSegment{
			Ticker:  "TEST",
			Candles: []models.Candle{},
		},
		Tickers:    []string{"SBER"},
		SearchFrom: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		SearchTo:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	results, err := scanner.Scan(query)

	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Scan() with empty segment returned %v results, expected 0", len(results))
	}
}

func TestScan_EmptyTickers(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	query := &chartmodels.ScanQuery{
		Segment: models.ChartSegment{
			Ticker:  "TEST",
			Candles: createTestCandles(10, 100.0, "up"),
		},
		Tickers:    []string{},
		SearchFrom: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		SearchTo:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	results, err := scanner.Scan(query)

	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Scan() with empty tickers returned %v results, expected 0", len(results))
	}
}

func TestScan_ShortSegment(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	mockFetcher.AddData("SBER", createTestCandles(100, 100.0, "up"))

	query := &chartmodels.ScanQuery{
		Segment: models.ChartSegment{
			Ticker:  "TEST",
			Candles: createTestCandles(2, 100.0, "up"),
		},
		Tickers:    []string{"SBER"},
		SearchFrom: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		SearchTo:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	results, err := scanner.Scan(query)

	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	t.Logf("Scan() with short segment returned %v results", len(results))
}

// Основная функциональность

func TestScan_ExactMatch(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	pattern := createTestCandles(20, 100.0, "up")
	mockFetcher.AddData("SBER", pattern)
	mockFetcher.AddData("GAZP", pattern)

	query := &chartmodels.ScanQuery{
		Segment: models.ChartSegment{
			Ticker:  "SBER",
			Candles: pattern[:10],
		},
		Tickers:    []string{"SBER", "GAZP"},
		SearchFrom: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		SearchTo:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		Options: chartmodels.ScanOptions{
			MinScale:  0.9,
			MaxScale:  1.1,
			Tolerance: 0.5,
		},
	}

	results, err := scanner.Scan(query)

	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(results) == 0 {
		t.Error("Scan() returned no results, expected at least one match")
	}
}

func TestScan_NoMatches(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	upPattern := createTestCandles(20, 100.0, "up")
	downPattern := createTestCandles(20, 100.0, "down")

	mockFetcher.AddData("SBER", downPattern)
	mockFetcher.AddData("GAZP", downPattern)

	query := &chartmodels.ScanQuery{
		Segment: models.ChartSegment{
			Ticker:  "TEST",
			Candles: upPattern[:10],
		},
		Tickers:    []string{"SBER", "GAZP"},
		SearchFrom: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		SearchTo:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		Options: chartmodels.ScanOptions{
			MinScale:  0.9,
			MaxScale:  1.1,
			Tolerance: 0.01,
		},
	}

	results, err := scanner.Scan(query)

	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(results) > 0 {
		t.Errorf("Scan() returned %v results with strict tolerance, expected 0", len(results))
	}
}

func TestScan_MultipleTickers(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	pattern := createTestCandles(30, 100.0, "volatile")
	mockFetcher.AddData("SBER", pattern)
	mockFetcher.AddData("GAZP", pattern)
	mockFetcher.AddData("LKOH", pattern)

	query := &chartmodels.ScanQuery{
		Segment: models.ChartSegment{
			Ticker:  "TEST",
			Candles: pattern[:15],
		},
		Tickers:    []string{"SBER", "GAZP", "LKOH"},
		SearchFrom: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		SearchTo:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		Options: chartmodels.ScanOptions{
			MinScale:  0.9,
			MaxScale:  1.1,
			Tolerance: 0.3,
		},
	}

	results, err := scanner.Scan(query)

	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	t.Logf("Scan() with 3 tickers returned %v results", len(results))
}

func TestScan_LongCandles(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	longPattern := createTestCandles(200, 100.0, "up")
	mockFetcher.AddData("SBER", longPattern)

	query := &chartmodels.ScanQuery{
		Segment: models.ChartSegment{
			Ticker:  "TEST",
			Candles: longPattern[:50],
		},
		Tickers:    []string{"SBER"},
		SearchFrom: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		SearchTo:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		Options: chartmodels.ScanOptions{
			MinScale:  0.9,
			MaxScale:  1.1,
			Tolerance: 0.3,
		},
	}

	results, err := scanner.Scan(query)

	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	t.Logf("Scan() with long candles returned %v results", len(results))
}

// Тестирование параметров сканирования

func TestScan_NarrowScale(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	pattern := createTestCandles(50, 100.0, "up")
	mockFetcher.AddData("SBER", pattern)

	query := &chartmodels.ScanQuery{
		Segment: models.ChartSegment{
			Ticker:  "TEST",
			Candles: pattern[:20],
		},
		Tickers:    []string{"SBER"},
		SearchFrom: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		SearchTo:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		Options: chartmodels.ScanOptions{
			MinScale:  0.95,
			MaxScale:  1.05,
			Tolerance: 0.5,
		},
	}

	results, err := scanner.Scan(query)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}

	t.Logf("narrow range: found %d matches", len(results))
}

func TestScan_WideScale(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	pattern := createTestCandles(50, 100.0, "up")
	mockFetcher.AddData("SBER", pattern)

	query := &chartmodels.ScanQuery{
		Segment: models.ChartSegment{
			Ticker:  "TEST",
			Candles: pattern[:20],
		},
		Tickers:    []string{"SBER"},
		SearchFrom: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		SearchTo:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		Options: chartmodels.ScanOptions{
			MinScale:  0.5,
			MaxScale:  2.0,
			Tolerance: 0.5,
		},
	}

	results, err := scanner.Scan(query)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}

	t.Logf("wide range: found %d matches", len(results))
}

func TestScan_ExactScale(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	pattern := createTestCandles(50, 100.0, "up")
	mockFetcher.AddData("SBER", pattern)

	query := &chartmodels.ScanQuery{
		Segment: models.ChartSegment{
			Ticker:  "TEST",
			Candles: pattern[:20],
		},
		Tickers:    []string{"SBER"},
		SearchFrom: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		SearchTo:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		Options: chartmodels.ScanOptions{
			MinScale:  1.0,
			MaxScale:  1.0,
			Tolerance: 0.5,
		},
	}

	results, err := scanner.Scan(query)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}

	t.Logf("exact match: found %d matches", len(results))
}

func TestScan_StrictTolerance(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	pattern := createTestCandles(30, 100.0, "up")
	mockFetcher.AddData("SBER", pattern)

	query := &chartmodels.ScanQuery{
		Segment: models.ChartSegment{
			Ticker:  "TEST",
			Candles: pattern[:15],
		},
		Tickers:    []string{"SBER"},
		SearchFrom: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		SearchTo:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		Options: chartmodels.ScanOptions{
			MinScale:  0.9,
			MaxScale:  1.1,
			Tolerance: 0.05,
		},
	}

	results, err := scanner.Scan(query)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}

	t.Logf("strict tolerance: found %d matches", len(results))
}

func TestScan_LooseTolerance(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	upPattern := createTestCandles(30, 100.0, "up")
	downPattern := createTestCandles(30, 100.0, "down")
	mockFetcher.AddData("SBER", downPattern)

	query := &chartmodels.ScanQuery{
		Segment: models.ChartSegment{
			Ticker:  "TEST",
			Candles: upPattern[:15],
		},
		Tickers:    []string{"SBER"},
		SearchFrom: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		SearchTo:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		Options: chartmodels.ScanOptions{
			MinScale:  0.9,
			MaxScale:  1.1,
			Tolerance: 0.9,
		},
	}

	results, err := scanner.Scan(query)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}

	t.Logf("loose tolerance: found %d matches", len(results))
}

func TestScan_DefaultOptions(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	pattern := createTestCandles(30, 100.0, "volatile")
	mockFetcher.AddData("SBER", pattern)

	query := &chartmodels.ScanQuery{
		Segment: models.ChartSegment{
			Ticker:  "TEST",
			Candles: pattern[:15],
		},
		Tickers:    []string{"SBER"},
		SearchFrom: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		SearchTo:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		// Options не установлены - должны примениться дефолтные значения
	}

	results, err := scanner.Scan(query)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}

	t.Logf("default options: found %d matches", len(results))
}
