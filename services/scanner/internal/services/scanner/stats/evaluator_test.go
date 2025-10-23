package stats

import (
	"errors"
	"testing"
	"time"

	"github.com/m1keee3/FinanceAnalyst/services/scanner/domain/models"
)

// MockFetcher для тестирования
type MockFetcher struct {
	fetchFunc func(ticker string, from, to time.Time) ([]models.Candle, error)
}

func (m *MockFetcher) Fetch(ticker string, from, to time.Time) ([]models.Candle, error) {
	if m.fetchFunc != nil {
		return m.fetchFunc(ticker, from, to)
	}
	return nil, nil
}

// Stats.Evaluator tests
// Граничные случаи

func TestComputeStats_NilScanner(t *testing.T) {
	var e *Evaluator
	matches := []models.ChartSegment{
		{
			Ticker: "AAPL",
			From:   time.Now(),
			To:     time.Now().Add(24 * time.Hour),
			Candles: []models.Candle{
				{Open: 100, Close: 110, High: 115, Low: 95},
			},
		},
	}

	stats, err := e.ComputeStats(matches, 5)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if stats == nil {
		t.Errorf("expected non-nil stats")
	}
	if stats.TotalMatches != 0 {
		t.Errorf("expected 0 total matches, got %d", stats.TotalMatches)
	}
}

func TestComputeStats_NilFetcher(t *testing.T) {
	e := &Evaluator{fetcher: nil}
	matches := []models.ChartSegment{
		{
			Ticker: "AAPL",
			From:   time.Now(),
			To:     time.Now().Add(24 * time.Hour),
			Candles: []models.Candle{
				{Open: 100, Close: 110, High: 115, Low: 95},
			},
		},
	}

	stats, err := e.ComputeStats(matches, 5)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if stats == nil {
		t.Errorf("expected non-nil stats")
	}
	if stats.TotalMatches != 0 {
		t.Errorf("expected 0 total matches, got %d", stats.TotalMatches)
	}
}

func TestComputeStats_EmptyMatches(t *testing.T) {
	mock := &MockFetcher{}
	e := NewEvaluator(mock)
	matches := []models.ChartSegment{}

	stats, err := e.ComputeStats(matches, 5)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if stats == nil {
		t.Errorf("expected non-nil stats")
	}
	if stats.TotalMatches != 0 {
		t.Errorf("expected 0 total matches, got %d", stats.TotalMatches)
	}
	if stats.PriceChange != 0 {
		t.Errorf("expected 0 price change, got %f", stats.PriceChange)
	}
	if stats.Probability != 0 {
		t.Errorf("expected 0 probability, got %f", stats.Probability)
	}
}

func TestComputeStats_FetcherError(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	mock := &MockFetcher{
		fetchFunc: func(ticker string, from, to time.Time) ([]models.Candle, error) {
			return nil, errors.New("fetch error")
		},
	}
	e := NewEvaluator(mock)
	matches := []models.ChartSegment{
		{
			Ticker: "AAPL",
			From:   baseDate,
			To:     baseDate.Add(24 * time.Hour),
			Candles: []models.Candle{
				{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95},
			},
		},
	}

	stats, err := e.ComputeStats(matches, 5)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	// При ошибках fetcher совпадения не учитываются
	if stats.TotalMatches != 0 {
		t.Errorf("expected 0 total matches, got %d", stats.TotalMatches)
	}
}

// Тестирование основной функциональности с фиксированным daysToWatch

func TestComputeStats_SingleMatch_PositiveChange(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	mock := &MockFetcher{
		fetchFunc: func(ticker string, from, to time.Time) ([]models.Candle, error) {
			// Возвращаем свечи с положительным изменением
			return []models.Candle{
				{Date: baseDate.Add(24 * time.Hour), Open: 100, Close: 105, High: 106, Low: 99},
				{Date: baseDate.Add(48 * time.Hour), Open: 105, Close: 110, High: 111, Low: 104},
				{Date: baseDate.Add(72 * time.Hour), Open: 110, Close: 115, High: 116, Low: 109},
			}, nil
		},
	}

	e := NewEvaluator(mock)
	matches := []models.ChartSegment{
		{
			Ticker: "AAPL",
			From:   baseDate,
			To:     baseDate.Add(12 * time.Hour),
			Candles: []models.Candle{
				{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95},
			},
		},
	}

	stats, err := e.ComputeStats(matches, 3)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if stats.TotalMatches != 1 {
		t.Errorf("expected 1 total match, got %d", stats.TotalMatches)
	}
	if stats.Probability != 1.0 {
		t.Errorf("expected probability 1.0, got %f", stats.Probability)
	}
	if stats.PriceChange <= 0 {
		t.Errorf("expected positive price change, got %f", stats.PriceChange)
	}
}

func TestComputeStats_SingleMatch_NegativeChange(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	mock := &MockFetcher{
		fetchFunc: func(ticker string, from, to time.Time) ([]models.Candle, error) {
			// Возвращаем свечи с отрицательным изменением
			return []models.Candle{
				{Date: baseDate.Add(24 * time.Hour), Open: 100, Close: 95, High: 101, Low: 94},
				{Date: baseDate.Add(48 * time.Hour), Open: 95, Close: 90, High: 96, Low: 89},
				{Date: baseDate.Add(72 * time.Hour), Open: 90, Close: 85, High: 91, Low: 84},
			}, nil
		},
	}

	e := NewEvaluator(mock)
	matches := []models.ChartSegment{
		{
			Ticker: "AAPL",
			From:   baseDate,
			To:     baseDate.Add(12 * time.Hour),
			Candles: []models.Candle{
				{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95},
			},
		},
	}

	stats, err := e.ComputeStats(matches, 3)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if stats.TotalMatches != 1 {
		t.Errorf("expected 1 total match, got %d", stats.TotalMatches)
	}
	if stats.Probability != 1.0 {
		t.Errorf("expected probability 1.0, got %f", stats.Probability)
	}
	if stats.PriceChange >= 0 {
		t.Errorf("expected negative price change, got %f", stats.PriceChange)
	}
}

func TestComputeStats_MultipleMatches_MixedChanges(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	callCount := 0
	mock := &MockFetcher{
		fetchFunc: func(ticker string, from, to time.Time) ([]models.Candle, error) {
			callCount++
			if callCount == 1 {
				// Первое совпадение: положительное изменение
				return []models.Candle{
					{Date: baseDate.Add(24 * time.Hour), Open: 100, Close: 105, High: 106, Low: 99},
					{Date: baseDate.Add(48 * time.Hour), Open: 105, Close: 110, High: 111, Low: 104},
				}, nil
			}
			// Второе совпадение: отрицательное изменение
			return []models.Candle{
				{Date: baseDate.Add(24 * time.Hour), Open: 100, Close: 95, High: 101, Low: 94},
				{Date: baseDate.Add(48 * time.Hour), Open: 95, Close: 90, High: 96, Low: 89},
			}, nil
		},
	}

	e := NewEvaluator(mock)
	matches := []models.ChartSegment{
		{
			Ticker: "AAPL",
			From:   baseDate,
			To:     baseDate.Add(12 * time.Hour),
			Candles: []models.Candle{
				{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95},
			},
		},
		{
			Ticker: "GOOGL",
			From:   baseDate,
			To:     baseDate.Add(12 * time.Hour),
			Candles: []models.Candle{
				{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95},
			},
		},
	}

	stats, err := e.ComputeStats(matches, 2)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if stats.TotalMatches != 2 {
		t.Errorf("expected 2 total matches, got %d", stats.TotalMatches)
	}
	if stats.Probability != 0.5 {
		t.Errorf("expected probability 0.5, got %f", stats.Probability)
	}
	if stats.PriceChange >= 0 {
		t.Errorf("expected negative price change (since negative trend is chosen when equal), got %f", stats.PriceChange)
	}
}

func TestComputeStats_MultipleMatches_AllPositive(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	mock := &MockFetcher{
		fetchFunc: func(ticker string, from, to time.Time) ([]models.Candle, error) {
			// Все совпадения: положительное изменение
			return []models.Candle{
				{Date: baseDate.Add(24 * time.Hour), Open: 100, Close: 105, High: 106, Low: 99},
				{Date: baseDate.Add(48 * time.Hour), Open: 105, Close: 110, High: 111, Low: 104},
			}, nil
		},
	}

	e := NewEvaluator(mock)
	matches := []models.ChartSegment{
		{
			Ticker: "AAPL",
			From:   baseDate,
			To:     baseDate.Add(12 * time.Hour),
			Candles: []models.Candle{
				{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95},
			},
		},
		{
			Ticker: "GOOGL",
			From:   baseDate,
			To:     baseDate.Add(12 * time.Hour),
			Candles: []models.Candle{
				{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95},
			},
		},
		{
			Ticker: "MSFT",
			From:   baseDate,
			To:     baseDate.Add(12 * time.Hour),
			Candles: []models.Candle{
				{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95},
			},
		},
	}

	stats, err := e.ComputeStats(matches, 2)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if stats.TotalMatches != 3 {
		t.Errorf("expected 3 total matches, got %d", stats.TotalMatches)
	}
	if stats.Probability != 1.0 {
		t.Errorf("expected probability 1.0, got %f", stats.Probability)
	}
	if stats.PriceChange <= 0 {
		t.Errorf("expected positive price change, got %f", stats.PriceChange)
	}
}

func TestComputeStats_NoDataAfterMatch(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	mock := &MockFetcher{
		fetchFunc: func(ticker string, from, to time.Time) ([]models.Candle, error) {
			// Нет данных после совпадения
			return []models.Candle{}, nil
		},
	}

	e := NewEvaluator(mock)
	matches := []models.ChartSegment{
		{
			Ticker: "AAPL",
			From:   baseDate,
			To:     baseDate.Add(12 * time.Hour),
			Candles: []models.Candle{
				{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95},
			},
		},
	}

	stats, err := e.ComputeStats(matches, 5)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if stats.TotalMatches != 0 {
		t.Errorf("expected 0 total matches, got %d", stats.TotalMatches)
	}
}

// Тестирование с daysToWatch = 0 (computeLineStats)

func TestComputeStats_LineStats_PositiveTrend(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	mock := &MockFetcher{
		fetchFunc: func(ticker string, from, to time.Time) ([]models.Candle, error) {
			// Возвращаем растущие свечи, затем падающую (должно остановиться)
			return []models.Candle{
				{Date: baseDate.Add(24 * time.Hour), Open: 100, Close: 105, High: 106, Low: 99},
				{Date: baseDate.Add(48 * time.Hour), Open: 105, Close: 110, High: 111, Low: 104},
				{Date: baseDate.Add(72 * time.Hour), Open: 110, Close: 115, High: 116, Low: 109},
				{Date: baseDate.Add(96 * time.Hour), Open: 115, Close: 110, High: 116, Low: 109}, // падающая - стоп
				{Date: baseDate.Add(120 * time.Hour), Open: 110, Close: 105, High: 111, Low: 104},
			}, nil
		},
	}

	e := NewEvaluator(mock)
	matches := []models.ChartSegment{
		{
			Ticker: "AAPL",
			From:   baseDate,
			To:     baseDate.Add(12 * time.Hour),
			Candles: []models.Candle{
				{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95},
			},
		},
	}

	stats, err := e.ComputeStats(matches, 0)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if stats.TotalMatches != 1 {
		t.Errorf("expected 1 total match, got %d", stats.TotalMatches)
	}
	if stats.Probability != 1.0 {
		t.Errorf("expected probability 1.0, got %f", stats.Probability)
	}
	if stats.PriceChange <= 0 {
		t.Errorf("expected positive price change, got %f", stats.PriceChange)
	}
}

func TestComputeStats_LineStats_NegativeTrend(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	mock := &MockFetcher{
		fetchFunc: func(ticker string, from, to time.Time) ([]models.Candle, error) {
			// Возвращаем падающие свечи, затем растущую (должно остановиться)
			return []models.Candle{
				{Date: baseDate.Add(24 * time.Hour), Open: 100, Close: 95, High: 101, Low: 94},
				{Date: baseDate.Add(48 * time.Hour), Open: 95, Close: 90, High: 96, Low: 89},
				{Date: baseDate.Add(72 * time.Hour), Open: 90, Close: 85, High: 91, Low: 84},
				{Date: baseDate.Add(96 * time.Hour), Open: 85, Close: 90, High: 91, Low: 84}, // растущая - стоп
				{Date: baseDate.Add(120 * time.Hour), Open: 90, Close: 95, High: 96, Low: 89},
			}, nil
		},
	}

	e := NewEvaluator(mock)
	matches := []models.ChartSegment{
		{
			Ticker: "AAPL",
			From:   baseDate,
			To:     baseDate.Add(12 * time.Hour),
			Candles: []models.Candle{
				{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95},
			},
		},
	}

	stats, err := e.ComputeStats(matches, 0)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if stats.TotalMatches != 1 {
		t.Errorf("expected 1 total match, got %d", stats.TotalMatches)
	}
	if stats.Probability != 1.0 {
		t.Errorf("expected probability 1.0, got %f", stats.Probability)
	}
	if stats.PriceChange >= 0 {
		t.Errorf("expected negative price change, got %f", stats.PriceChange)
	}
}

func TestComputeStats_LineStats_EmptyData(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	mock := &MockFetcher{
		fetchFunc: func(ticker string, from, to time.Time) ([]models.Candle, error) {
			return []models.Candle{}, nil
		},
	}

	e := NewEvaluator(mock)
	matches := []models.ChartSegment{
		{
			Ticker: "AAPL",
			From:   baseDate,
			To:     baseDate.Add(12 * time.Hour),
			Candles: []models.Candle{
				{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95},
			},
		},
	}

	stats, err := e.ComputeStats(matches, 0)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if stats.TotalMatches != 0 {
		t.Errorf("expected 0 total matches, got %d", stats.TotalMatches)
	}
}

func TestComputeStats_LineStats_MultipleMatches(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	callCount := 0
	mock := &MockFetcher{
		fetchFunc: func(ticker string, from, to time.Time) ([]models.Candle, error) {
			callCount++
			if callCount%2 == 1 {
				// Нечетные: положительный тренд
				return []models.Candle{
					{Date: baseDate.Add(24 * time.Hour), Open: 100, Close: 105, High: 106, Low: 99},
					{Date: baseDate.Add(48 * time.Hour), Open: 105, Close: 110, High: 111, Low: 104},
				}, nil
			}
			// Четные: отрицательный тренд
			return []models.Candle{
				{Date: baseDate.Add(24 * time.Hour), Open: 100, Close: 95, High: 101, Low: 94},
				{Date: baseDate.Add(48 * time.Hour), Open: 95, Close: 90, High: 96, Low: 89},
			}, nil
		},
	}

	e := NewEvaluator(mock)
	matches := []models.ChartSegment{
		{Ticker: "AAPL", From: baseDate, To: baseDate.Add(12 * time.Hour),
			Candles: []models.Candle{{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95}}},
		{Ticker: "GOOGL", From: baseDate, To: baseDate.Add(12 * time.Hour),
			Candles: []models.Candle{{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95}}},
		{Ticker: "MSFT", From: baseDate, To: baseDate.Add(12 * time.Hour),
			Candles: []models.Candle{{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95}}},
	}

	stats, err := e.ComputeStats(matches, 0)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if stats.TotalMatches != 3 {
		t.Errorf("expected 3 total matches, got %d", stats.TotalMatches)
	}
	if stats.Probability <= 0 || stats.Probability > 1 {
		t.Errorf("expected probability between 0 and 1, got %f", stats.Probability)
	}
}
