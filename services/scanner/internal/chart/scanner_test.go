package chart

import (
	"testing"
	"time"

	"github.com/m1keee3/FinanceAnalyst/services/scanner/domain/models"
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

func TestFindMatches_NilScanner(t *testing.T) {
	var scanner *Scanner

	segment := models.ChartSegment{
		Candles: createTestCandles(10, 100.0, "up"),
	}

	results, err := scanner.FindMatches(
		segment,
		[]string{"SBER"},
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		nil,
	)

	if err != nil {
		t.Errorf("FindMatches() error = %v, want nil", err)
	}
	if results != nil {
		t.Errorf("FindMatches() returned %v, want nil", results)
	}
}

func TestFindMatches_NilFetcher(t *testing.T) {
	scanner := &Scanner{fetcher: nil}

	segment := models.ChartSegment{
		Candles: createTestCandles(10, 100.0, "up"),
	}

	results, err := scanner.FindMatches(
		segment,
		[]string{"SBER"},
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		nil,
	)

	if err != nil {
		t.Errorf("FindMatches() error = %v, want nil", err)
	}
	if results != nil {
		t.Errorf("FindMatches() returned %v, want nil", results)
	}
}

func TestFindMatches_EmptySegment(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	segment := models.ChartSegment{
		Ticker:  "TEST",
		Candles: []models.Candle{},
	}

	results, err := scanner.FindMatches(
		segment,
		[]string{"SBER"},
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		nil,
	)

	if err != nil {
		t.Fatalf("FindMatches() error = %v", err)
	}

	if len(results) != 0 {
		t.Errorf("FindMatches() with empty segment returned %v results, expected 0", len(results))
	}
}

func TestFindMatches_EmptyTickers(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	segment := models.ChartSegment{
		Ticker:  "TEST",
		Candles: createTestCandles(10, 100.0, "up"),
	}

	results, err := scanner.FindMatches(
		segment,
		[]string{},
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		nil,
	)

	if err != nil {
		t.Fatalf("FindMatches() error = %v", err)
	}

	if len(results) != 0 {
		t.Errorf("FindMatches() with empty tickers returned %v results, expected 0", len(results))
	}
}

func TestFindMatches_ShortSegment(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	mockFetcher.AddData("SBER", createTestCandles(100, 100.0, "up"))

	segment := models.ChartSegment{
		Ticker:  "TEST",
		Candles: createTestCandles(2, 100.0, "up"),
	}

	results, err := scanner.FindMatches(
		segment,
		[]string{"SBER"},
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		nil,
	)

	if err != nil {
		t.Fatalf("FindMatches() error = %v", err)
	}

	t.Logf("FindMatches() with short segment returned %v results", len(results))
}

// Основная функциональность

func TestFindMatches_ExactMatch(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	pattern := createTestCandles(20, 100.0, "up")
	mockFetcher.AddData("SBER", pattern)
	mockFetcher.AddData("GAZP", pattern)

	segment := models.ChartSegment{
		Ticker:  "SBER",
		Candles: pattern[:10],
	}

	options := &ScanOptions{
		MinScale:  0.9,
		MaxScale:  1.1,
		Tolerance: 0.5,
	}

	results, err := scanner.FindMatches(
		segment,
		[]string{"SBER", "GAZP"},
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		options,
	)

	if err != nil {
		t.Fatalf("FindMatches() error = %v", err)
	}

	if len(results) == 0 {
		t.Error("FindMatches() returned no results, expected at least one match")
	}
}

func TestFindMatches_NoMatches(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	upPattern := createTestCandles(20, 100.0, "up")
	downPattern := createTestCandles(20, 100.0, "down")

	mockFetcher.AddData("SBER", downPattern)
	mockFetcher.AddData("GAZP", downPattern)

	segment := models.ChartSegment{
		Ticker:  "TEST",
		Candles: upPattern[:10],
	}

	options := &ScanOptions{
		MinScale:  0.9,
		MaxScale:  1.1,
		Tolerance: 0.01,
	}

	results, err := scanner.FindMatches(
		segment,
		[]string{"SBER", "GAZP"},
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		options,
	)

	if err != nil {
		t.Fatalf("FindMatches() error = %v", err)
	}

	if len(results) > 0 {
		t.Errorf("FindMatches() returned %v results with strict tolerance, expected 0", len(results))
	}
}

func TestFindMatches_MultipleTickers(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	pattern := createTestCandles(30, 100.0, "volatile")
	mockFetcher.AddData("SBER", pattern)
	mockFetcher.AddData("GAZP", pattern)
	mockFetcher.AddData("LKOH", pattern)

	segment := models.ChartSegment{
		Ticker:  "TEST",
		Candles: pattern[:15],
	}

	options := &ScanOptions{
		MinScale:  0.9,
		MaxScale:  1.1,
		Tolerance: 0.3,
	}

	results, err := scanner.FindMatches(
		segment,
		[]string{"SBER", "GAZP", "LKOH"},
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		options,
	)

	if err != nil {
		t.Fatalf("FindMatches() error = %v", err)
	}

	t.Logf("FindMatches() with 3 tickers returned %v results", len(results))
}

func TestFindMatches_LongCandles(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	longPattern := createTestCandles(200, 100.0, "up")
	mockFetcher.AddData("SBER", longPattern)

	segment := models.ChartSegment{
		Ticker:  "TEST",
		Candles: longPattern[:50],
	}

	options := &ScanOptions{
		MinScale:  0.9,
		MaxScale:  1.1,
		Tolerance: 0.3,
	}

	results, err := scanner.FindMatches(
		segment,
		[]string{"SBER"},
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		options,
	)

	if err != nil {
		t.Fatalf("FindMatches() error = %v", err)
	}

	t.Logf("FindMatches() with long candles returned %v results", len(results))
}

// Тестирование параметров сканирования

func TestFindMatches_NarrowScale(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	pattern := createTestCandles(50, 100.0, "up")
	mockFetcher.AddData("SBER", pattern)

	segment := models.ChartSegment{
		Ticker:  "TEST",
		Candles: pattern[:20],
	}

	options := &ScanOptions{
		MinScale:  0.95,
		MaxScale:  1.05,
		Tolerance: 0.5,
	}

	results, err := scanner.FindMatches(
		segment,
		[]string{"SBER"},
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		options,
	)

	if err != nil {
		t.Errorf("FindMatches() error = %v", err)
	}

	t.Logf("narrow range: found %d matches", len(results))
}

func TestFindMatches_WideScale(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	pattern := createTestCandles(50, 100.0, "up")
	mockFetcher.AddData("SBER", pattern)

	segment := models.ChartSegment{
		Ticker:  "TEST",
		Candles: pattern[:20],
	}

	options := &ScanOptions{
		MinScale:  0.5,
		MaxScale:  2.0,
		Tolerance: 0.5,
	}

	results, err := scanner.FindMatches(
		segment,
		[]string{"SBER"},
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		options,
	)

	if err != nil {
		t.Errorf("FindMatches() error = %v", err)
	}

	t.Logf("wide range: found %d matches", len(results))
}

func TestFindMatches_ExactScale(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	pattern := createTestCandles(50, 100.0, "up")
	mockFetcher.AddData("SBER", pattern)

	segment := models.ChartSegment{
		Ticker:  "TEST",
		Candles: pattern[:20],
	}

	options := &ScanOptions{
		MinScale:  1.0,
		MaxScale:  1.0,
		Tolerance: 0.5,
	}

	results, err := scanner.FindMatches(
		segment,
		[]string{"SBER"},
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		options,
	)

	if err != nil {
		t.Errorf("FindMatches() error = %v", err)
	}

	t.Logf("exact match: found %d matches", len(results))
}

func TestFindMatches_StrictTolerance(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	pattern := createTestCandles(30, 100.0, "up")
	mockFetcher.AddData("SBER", pattern)

	segment := models.ChartSegment{
		Ticker:  "TEST",
		Candles: pattern[:15],
	}

	options := &ScanOptions{
		MinScale:  0.9,
		MaxScale:  1.1,
		Tolerance: 0.05,
	}

	results, err := scanner.FindMatches(
		segment,
		[]string{"SBER"},
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		options,
	)

	if err != nil {
		t.Errorf("FindMatches() error = %v", err)
	}

	t.Logf("strict tolerance: found %d matches", len(results))
}

func TestFindMatches_LooseTolerance(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	upPattern := createTestCandles(30, 100.0, "up")
	downPattern := createTestCandles(30, 100.0, "down")
	mockFetcher.AddData("SBER", downPattern)

	segment := models.ChartSegment{
		Ticker:  "TEST",
		Candles: upPattern[:15],
	}

	options := &ScanOptions{
		MinScale:  0.9,
		MaxScale:  1.1,
		Tolerance: 0.9,
	}

	results, err := scanner.FindMatches(
		segment,
		[]string{"SBER"},
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		options,
	)

	if err != nil {
		t.Errorf("FindMatches() error = %v", err)
	}

	t.Logf("loose tolerance: found %d matches", len(results))
}

func TestFindMatches_DefaultOptions(t *testing.T) {
	mockFetcher := NewMockFetcher()
	scanner := NewScanner(mockFetcher)

	pattern := createTestCandles(30, 100.0, "volatile")
	mockFetcher.AddData("SBER", pattern)

	segment := models.ChartSegment{
		Ticker:  "TEST",
		Candles: pattern[:15],
	}

	results, err := scanner.FindMatches(
		segment,
		[]string{"SBER"},
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		nil,
	)

	if err != nil {
		t.Errorf("FindMatches() error = %v", err)
	}

	t.Logf("default options: found %d matches", len(results))
}
