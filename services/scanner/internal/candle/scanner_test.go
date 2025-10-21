package candle

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

// Scan tests
// Граничные случаи

func TestScan_NilScanner(t *testing.T) {
	var s *Scanner
	query := &ScanQuery{
		Segment: models.ChartSegment{
			Candles: []models.Candle{
				{Open: 100, Close: 110, High: 115, Low: 95},
			},
		},
		Tickers: []string{"AAPL"},
	}

	matches, err := s.Scan(query)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if matches != nil {
		t.Errorf("expected nil matches, got %v", matches)
	}
}

func TestScan_NilFetcher(t *testing.T) {
	s := &Scanner{fetcher: nil}
	query := &ScanQuery{
		Segment: models.ChartSegment{
			Candles: []models.Candle{
				{Open: 100, Close: 110, High: 115, Low: 95},
			},
		},
		Tickers: []string{"AAPL"},
	}

	matches, err := s.Scan(query)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if matches != nil {
		t.Errorf("expected nil matches, got %v", matches)
	}
}

func TestScan_EmptySegment(t *testing.T) {
	mock := &MockFetcher{}
	s := NewScanner(mock)
	query := &ScanQuery{
		Segment: models.ChartSegment{
			Candles: []models.Candle{},
		},
		Tickers: []string{"AAPL"},
	}

	matches, err := s.Scan(query)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if matches != nil {
		t.Errorf("expected nil matches, got %v", matches)
	}
}

func TestScan_EmptyTickers(t *testing.T) {
	mock := &MockFetcher{}
	s := NewScanner(mock)
	query := &ScanQuery{
		Segment: models.ChartSegment{
			Candles: []models.Candle{
				{Open: 100, Close: 110, High: 115, Low: 95},
			},
		},
		Tickers: []string{},
	}

	matches, err := s.Scan(query)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if matches != nil {
		t.Errorf("expected nil matches, got %v", matches)
	}
}

func TestScan_FetcherError(t *testing.T) {
	mock := &MockFetcher{
		fetchFunc: func(ticker string, from, to time.Time) ([]models.Candle, error) {
			return nil, errors.New("fetch error")
		},
	}
	s := NewScanner(mock)
	query := &ScanQuery{
		Segment: models.ChartSegment{
			Candles: []models.Candle{
				{Open: 100, Close: 110, High: 115, Low: 95},
			},
		},
		Tickers: []string{"AAPL"},
	}

	matches, err := s.Scan(query)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

// Тестирование основной функциональности

func TestScan_NoMatches(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	mock := &MockFetcher{
		fetchFunc: func(ticker string, from, to time.Time) ([]models.Candle, error) {
			// Возвращаем свечи, которые не совпадают с сегментом
			return []models.Candle{
				{Date: baseDate, Open: 200, Close: 180, High: 210, Low: 170},
				{Date: baseDate.Add(24 * time.Hour), Open: 180, Close: 160, High: 190, Low: 150},
			}, nil
		},
	}

	s := NewScanner(mock)
	query := &ScanQuery{
		Segment: models.ChartSegment{
			Candles: []models.Candle{
				{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95},
				{Date: baseDate.Add(24 * time.Hour), Open: 110, Close: 120, High: 125, Low: 105},
			},
		},
		Tickers:    []string{"AAPL"},
		SearchFrom: baseDate,
		SearchTo:   baseDate.Add(48 * time.Hour),
	}

	matches, err := s.Scan(query)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

func TestScan_ExactMatch(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Создаем сегмент из двух свечей
	pattern := []models.Candle{
		{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95},
		{Date: baseDate.Add(24 * time.Hour), Open: 110, Close: 120, High: 125, Low: 105},
	}

	mock := &MockFetcher{
		fetchFunc: func(ticker string, from, to time.Time) ([]models.Candle, error) {
			// Возвращаем данные с точным совпадением сегмента в середине
			return []models.Candle{
				{Date: baseDate, Open: 50, Close: 60, High: 65, Low: 45},
				{Date: baseDate.Add(24 * time.Hour), Open: 100, Close: 110, High: 115, Low: 95},
				{Date: baseDate.Add(48 * time.Hour), Open: 110, Close: 120, High: 125, Low: 105},
				{Date: baseDate.Add(72 * time.Hour), Open: 200, Close: 180, High: 210, Low: 170},
			}, nil
		},
	}

	s := NewScanner(mock)
	query := &ScanQuery{
		Segment: models.ChartSegment{
			Candles: pattern,
		},
		Tickers:    []string{"AAPL"},
		SearchFrom: baseDate,
		SearchTo:   baseDate.Add(120 * time.Hour),
		Options: ScanOptions{
			TailLen:         0,
			BodyTolerance:   0.01,
			ShadowTolerance: 0.01,
		},
	}

	matches, err := s.Scan(query)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(matches) != 1 {
		t.Errorf("expected 1 match, got %d", len(matches))
	}
	if len(matches) > 0 {
		if matches[0].Ticker != "AAPL" {
			t.Errorf("expected ticker AAPL, got %s", matches[0].Ticker)
		}
		if len(matches[0].Candles) != 2 {
			t.Errorf("expected 2 candles in match, got %d", len(matches[0].Candles))
		}
	}
}

func TestScan_MultipleMatches(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	pattern := []models.Candle{
		{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95},
	}

	mock := &MockFetcher{
		fetchFunc: func(ticker string, from, to time.Time) ([]models.Candle, error) {
			// Возвращаем данные с несколькими совпадениями
			return []models.Candle{
				{Date: baseDate.Add(24 * time.Hour), Open: 100, Close: 110, High: 115, Low: 95},
				{Date: baseDate.Add(48 * time.Hour), Open: 200, Close: 180, High: 210, Low: 170},
				{Date: baseDate.Add(72 * time.Hour), Open: 100, Close: 110, High: 115, Low: 95},
			}, nil
		},
	}

	s := NewScanner(mock)
	query := &ScanQuery{
		Segment: models.ChartSegment{
			Candles: pattern,
		},
		Tickers:    []string{"AAPL"},
		SearchFrom: baseDate,
		SearchTo:   baseDate.Add(96 * time.Hour),
		Options: ScanOptions{
			TailLen:         0,
			BodyTolerance:   0.01,
			ShadowTolerance: 0.01,
		},
	}

	matches, err := s.Scan(query)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(matches))
	}
}

func TestScan_MultipleTickers(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	pattern := []models.Candle{
		{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95},
	}

	mock := &MockFetcher{
		fetchFunc: func(ticker string, from, to time.Time) ([]models.Candle, error) {
			// Для каждого тикера возвращаем одно совпадение
			return []models.Candle{
				{Date: baseDate.Add(24 * time.Hour), Open: 100, Close: 110, High: 115, Low: 95},
			}, nil
		},
	}

	s := NewScanner(mock)
	query := &ScanQuery{
		Segment: models.ChartSegment{
			Candles: pattern,
		},
		Tickers:    []string{"AAPL", "GOOGL", "MSFT"},
		SearchFrom: baseDate,
		SearchTo:   baseDate.Add(48 * time.Hour),
		Options: ScanOptions{
			TailLen:         0,
			BodyTolerance:   0.01,
			ShadowTolerance: 0.01,
		},
	}

	matches, err := s.Scan(query)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(matches) != 3 {
		t.Errorf("expected 3 matches (one per ticker), got %d", len(matches))
	}

	// Проверяем, что все тикеры представлены
	tickerMap := make(map[string]bool)
	for _, match := range matches {
		tickerMap[match.Ticker] = true
	}
	for _, ticker := range query.Tickers {
		if !tickerMap[ticker] {
			t.Errorf("expected to find ticker %s in matches", ticker)
		}
	}
}

// Тестирование параметров сканирования

func TestScan_WithTailLen(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Паттерн с хвостом (первая свеча падает, вторая растет)
	pattern := []models.Candle{
		{Date: baseDate, Open: 110, Close: 100, High: 115, Low: 95},                     // падающая (хвост)
		{Date: baseDate.Add(24 * time.Hour), Open: 100, Close: 120, High: 125, Low: 95}, // растущая
	}

	mock := &MockFetcher{
		fetchFunc: func(ticker string, from, to time.Time) ([]models.Candle, error) {
			// Совпадение: падающая + растущая
			return []models.Candle{
				{Date: baseDate.Add(48 * time.Hour), Open: 110, Close: 100, High: 115, Low: 95},
				{Date: baseDate.Add(72 * time.Hour), Open: 100, Close: 120, High: 125, Low: 95},
				// Несовпадение: растущая + растущая (неправильный знак хвоста)
				{Date: baseDate.Add(96 * time.Hour), Open: 100, Close: 110, High: 115, Low: 95},
				{Date: baseDate.Add(120 * time.Hour), Open: 110, Close: 130, High: 135, Low: 105},
			}, nil
		},
	}

	s := NewScanner(mock)
	query := &ScanQuery{
		Segment: models.ChartSegment{
			Candles: pattern,
		},
		Tickers:    []string{"AAPL"},
		SearchFrom: baseDate,
		SearchTo:   baseDate.Add(144 * time.Hour),
		Options: ScanOptions{
			TailLen:         1, // первая свеча - хвост
			BodyTolerance:   0.01,
			ShadowTolerance: 0.01,
		},
	}

	matches, err := s.Scan(query)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(matches) != 1 {
		t.Errorf("expected 1 match (with correct tail sign), got %d", len(matches))
	}
}

func TestScan_WithTolerance(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	pattern := []models.Candle{
		{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95},
	}

	mock := &MockFetcher{
		fetchFunc: func(ticker string, from, to time.Time) ([]models.Candle, error) {
			// Возвращаем свечу с небольшим отклонением (после нормализации)
			// После нормализации Open=100, Close=110 становятся пропорциональными
			// Эта свеча должна попасть в tolerance
			return []models.Candle{
				{Date: baseDate.Add(24 * time.Hour), Open: 100, Close: 110, High: 115, Low: 95},
			}, nil
		},
	}

	s := NewScanner(mock)
	query := &ScanQuery{
		Segment: models.ChartSegment{
			Candles: pattern,
		},
		Tickers:    []string{"AAPL"},
		SearchFrom: baseDate,
		SearchTo:   baseDate.Add(48 * time.Hour),
		Options: ScanOptions{
			TailLen:         0,
			BodyTolerance:   0.1,
			ShadowTolerance: 0.1,
		},
	}

	matches, err := s.Scan(query)
	if err != nil {
		t.Errorf("expected no error with high tolerance, got %v", err)
	}
	if len(matches) == 0 {
		t.Errorf("expected at least 1 match with high tolerance, got %d", len(matches))
	}
}

func TestScan_DefaultOptions(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	pattern := []models.Candle{
		{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95},
	}

	mock := &MockFetcher{
		fetchFunc: func(ticker string, from, to time.Time) ([]models.Candle, error) {
			return []models.Candle{
				{Date: baseDate.Add(24 * time.Hour), Open: 100, Close: 110, High: 115, Low: 95},
			}, nil
		},
	}

	s := NewScanner(mock)
	query := &ScanQuery{
		Segment: models.ChartSegment{
			Candles: pattern,
		},
		Tickers:    []string{"AAPL"},
		SearchFrom: baseDate,
		SearchTo:   baseDate.Add(48 * time.Hour),
		// Options не установлены - должны примениться дефолтные значения
	}

	matches, err := s.Scan(query)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if matches == nil {
		t.Errorf("expected non-nil matches slice")
	}
}

func TestScan_TailLenGreaterThanSegment(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	pattern := []models.Candle{
		{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95},
	}

	mock := &MockFetcher{
		fetchFunc: func(ticker string, from, to time.Time) ([]models.Candle, error) {
			return []models.Candle{
				{Date: baseDate.Add(24 * time.Hour), Open: 100, Close: 110, High: 115, Low: 95},
			}, nil
		},
	}

	s := NewScanner(mock)
	query := &ScanQuery{
		Segment: models.ChartSegment{
			Candles: pattern,
		},
		Tickers:    []string{"AAPL"},
		SearchFrom: baseDate,
		SearchTo:   baseDate.Add(48 * time.Hour),
		Options: ScanOptions{
			TailLen:         10,
			BodyTolerance:   0.1,
			ShadowTolerance: 0.1,
		},
	}

	matches, err := s.Scan(query)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

func TestScan_NegativeTailLen(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	pattern := []models.Candle{
		{Date: baseDate, Open: 100, Close: 110, High: 115, Low: 95},
	}

	mock := &MockFetcher{
		fetchFunc: func(ticker string, from, to time.Time) ([]models.Candle, error) {
			return []models.Candle{
				{Date: baseDate.Add(24 * time.Hour), Open: 100, Close: 110, High: 115, Low: 95},
			}, nil
		},
	}

	s := NewScanner(mock)
	query := &ScanQuery{
		Segment: models.ChartSegment{
			Candles: pattern,
		},
		Tickers:    []string{"AAPL"},
		SearchFrom: baseDate,
		SearchTo:   baseDate.Add(48 * time.Hour),
		Options: ScanOptions{
			TailLen:         -5,
			BodyTolerance:   0.1,
			ShadowTolerance: 0.1,
		},
	}

	matches, err := s.Scan(query)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	// Функция должна скорректировать TailLen до 0 и работать нормально
	if len(matches) == 0 {
		t.Errorf("expected at least 1 match, got 0")
	}
}

// IsOverlap tests (остаются без изменений)

func TestIsOverlap_DifferentTickers(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	seg1 := models.ChartSegment{
		Ticker: "AAPL",
		From:   baseDate,
		To:     baseDate.Add(24 * time.Hour),
	}

	seg2 := models.ChartSegment{
		Ticker: "GOOGL",
		From:   baseDate,
		To:     baseDate.Add(24 * time.Hour),
	}

	if IsOverlap(seg1, seg2) {
		t.Error("expected no overlap for different tickers")
	}
}

func TestIsOverlap_NoOverlap_Seg1BeforeSeg2(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	seg1 := models.ChartSegment{
		Ticker: "AAPL",
		From:   baseDate,
		To:     baseDate.Add(24 * time.Hour), // 1 янв - 2 янв
	}

	seg2 := models.ChartSegment{
		Ticker: "AAPL",
		From:   baseDate.Add(72 * time.Hour), // 4 янв
		To:     baseDate.Add(96 * time.Hour), // 5 янв
	}

	if IsOverlap(seg1, seg2) {
		t.Error("expected no overlap when seg1 ends before seg2 starts")
	}
}

func TestIsOverlap_NoOverlap_Seg2BeforeSeg1(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	seg1 := models.ChartSegment{
		Ticker: "AAPL",
		From:   baseDate.Add(72 * time.Hour), // 4 янв
		To:     baseDate.Add(96 * time.Hour), // 5 янв
	}

	seg2 := models.ChartSegment{
		Ticker: "AAPL",
		From:   baseDate,                     // 1 янв
		To:     baseDate.Add(24 * time.Hour), // 2 янв
	}

	if IsOverlap(seg1, seg2) {
		t.Error("expected no overlap when seg2 ends before seg1 starts")
	}
}

func TestIsOverlap_PartialOverlap(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	seg1 := models.ChartSegment{
		Ticker: "AAPL",
		From:   baseDate,                     // 1 янв
		To:     baseDate.Add(72 * time.Hour), // 4 янв
	}

	seg2 := models.ChartSegment{
		Ticker: "AAPL",
		From:   baseDate.Add(48 * time.Hour), // 3 янв
		To:     baseDate.Add(96 * time.Hour), // 5 янв
	}

	if !IsOverlap(seg1, seg2) {
		t.Error("expected overlap for partially overlapping segments")
	}
}

func TestIsOverlap_CompleteOverlap_Seg1ContainsSeg2(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	seg1 := models.ChartSegment{
		Ticker: "AAPL",
		From:   baseDate,                      // 1 янв
		To:     baseDate.Add(120 * time.Hour), // 6 янв
	}

	seg2 := models.ChartSegment{
		Ticker: "AAPL",
		From:   baseDate.Add(24 * time.Hour), // 2 янв
		To:     baseDate.Add(72 * time.Hour), // 4 янв
	}

	if !IsOverlap(seg1, seg2) {
		t.Error("expected overlap when seg1 completely contains seg2")
	}
}

func TestIsOverlap_CompleteOverlap_Seg2ContainsSeg1(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	seg1 := models.ChartSegment{
		Ticker: "AAPL",
		From:   baseDate.Add(24 * time.Hour), // 2 янв
		To:     baseDate.Add(72 * time.Hour), // 4 янв
	}

	seg2 := models.ChartSegment{
		Ticker: "AAPL",
		From:   baseDate,                      // 1 янв
		To:     baseDate.Add(120 * time.Hour), // 6 янв
	}

	if !IsOverlap(seg1, seg2) {
		t.Error("expected overlap when seg2 completely contains seg1")
	}
}

func TestIsOverlap_IdenticalSegments(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	seg1 := models.ChartSegment{
		Ticker: "AAPL",
		From:   baseDate,
		To:     baseDate.Add(72 * time.Hour),
	}

	seg2 := models.ChartSegment{
		Ticker: "AAPL",
		From:   baseDate,
		To:     baseDate.Add(72 * time.Hour),
	}

	if !IsOverlap(seg1, seg2) {
		t.Error("expected overlap for identical segments")
	}
}

func TestIsOverlap_TouchingBoundaries_NoOverlap(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// seg1 заканчивается ровно когда seg2 начинается
	seg1 := models.ChartSegment{
		Ticker: "AAPL",
		From:   baseDate,
		To:     baseDate.Add(24 * time.Hour), // 2 янв 00:00
	}

	seg2 := models.ChartSegment{
		Ticker: "AAPL",
		From:   baseDate.Add(24 * time.Hour), // 2 янв 00:00
		To:     baseDate.Add(48 * time.Hour), // 3 янв 00:00
	}

	// Граничный случай: когда To одного равно From другого
	// Поскольку используется Before (строгое сравнение), это считается касанием, но не наложением
	if IsOverlap(seg1, seg2) {
		t.Error("expected no overlap when segments only touch at boundary")
	}
}

func TestIsOverlap_SinglePointOverlap(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Сегменты пересекаются на одну минуту
	seg1 := models.ChartSegment{
		Ticker: "AAPL",
		From:   baseDate,
		To:     baseDate.Add(24*time.Hour + 1*time.Minute), // 2 янв 00:01
	}

	seg2 := models.ChartSegment{
		Ticker: "AAPL",
		From:   baseDate.Add(24 * time.Hour), // 2 янв 00:00
		To:     baseDate.Add(48 * time.Hour), // 3 янв 00:00
	}

	if !IsOverlap(seg1, seg2) {
		t.Error("expected overlap even for minimal time overlap")
	}
}

func TestIsOverlap_DifferentTickersSameTime(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	seg1 := models.ChartSegment{
		Ticker: "AAPL",
		From:   baseDate,
		To:     baseDate.Add(72 * time.Hour),
	}

	seg2 := models.ChartSegment{
		Ticker: "GOOGL",
		From:   baseDate.Add(24 * time.Hour),
		To:     baseDate.Add(48 * time.Hour),
	}

	if IsOverlap(seg1, seg2) {
		t.Error("expected no overlap for different tickers even with time overlap")
	}
}
