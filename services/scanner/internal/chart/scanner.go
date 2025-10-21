package chart

import (
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/m1keee3/FinanceAnalyst/services/scanner/domain/models"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/pkg/utils"
)

type Fetcher interface {
	Fetch(ticker string, from, to time.Time) ([]models.Candle, error)
}

type Scanner struct {
	fetcher Fetcher
}

func NewScanner(fetcher Fetcher) *Scanner {
	return &Scanner{
		fetcher: fetcher,
	}
}

// ScanOptions определяет параметры сканирования графиков с использованием DTW
type ScanOptions struct {
	// MinScale минимальная длина сегмента относительно входного
	MinScale float64
	// MaxScale максимальная длина сегмента относительно входного
	MaxScale float64
	// Tolerance допустимая степень отличия от 0 до 1, где 0 = идентичные графики, 1 = максимальное отличие
	Tolerance float64
}

func (o *ScanOptions) withDefaults() ScanOptions {
	out := ScanOptions{
		MinScale:  0.75,
		MaxScale:  1.5,
		Tolerance: 0.1,
	}
	if o == nil {
		return out
	}
	if o.MinScale > 0 {
		out.MinScale = o.MinScale
	}
	if o.MaxScale > 0 {
		out.MaxScale = o.MaxScale
	}
	if o.Tolerance > 0 && o.Tolerance <= 1.0 {
		out.Tolerance = o.Tolerance
	}
	return out
}

// match представляет найденное совпадение с метрикой качества
type match struct {
	Segment  models.ChartSegment
	Distance float64 // Нормализованное DTW расстояние от 0 (идентично) до 1 (максимальное отличие)
}

// FindMatches ищет похожие паттерны в данных тикеров используя DTW алгоритм
func (s *Scanner) FindMatches(segment models.ChartSegment, tickers []string, searchFrom, searchTo time.Time, options *ScanOptions) ([]models.ChartSegment, error) {
	if s == nil || s.fetcher == nil {
		return nil, nil
	}

	if len(segment.Candles) == 0 || len(tickers) == 0 {
		return nil, nil
	}

	opts := options.withDefaults()
	seedLen := len(segment.Candles)

	minLen := int(float64(seedLen) * opts.MinScale)
	maxLen := int(float64(seedLen) * opts.MaxScale)
	if minLen < 1 {
		minLen = 1
	}

	seedVec := getPricesVec(segment.Candles, len(segment.Candles)*2)

	resampledLength := len(seedVec)

	var allMatches []match
	var mu sync.Mutex

	// Параллельная обработка тикеров
	tickerCh := make(chan string, len(tickers))
	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for ticker := range tickerCh {
			candles, err := s.fetcher.Fetch(ticker, searchFrom, searchTo)
			if err != nil {
				continue
			}

			if len(candles) < minLen {
				continue
			}

			matches := s.findMatches(seedVec, ticker, candles, minLen, maxLen, opts.Tolerance, resampledLength)

			mu.Lock()
			allMatches = append(allMatches, matches...)
			mu.Unlock()
		}
	}

	numWorkers := runtime.NumCPU()
	if numWorkers > len(tickers) {
		numWorkers = len(tickers)
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker()
	}

	for _, ticker := range tickers {
		tickerCh <- ticker
	}
	close(tickerCh)

	wg.Wait()

	allMatches = removeOverlaps(allMatches)

	result := make([]models.ChartSegment, len(allMatches))
	for i, m := range allMatches {
		result[i] = m.Segment
	}

	return result, nil
}

// findMatches ищет совпадения для заданного seed вектора в массиве свечей
// с учетом диапазона длин от minLen до maxLen
func (s *Scanner) findMatches(seedVec []float64, ticker string, candles []models.Candle, minLen, maxLen int, tolerance float64, resampledLength int) []match {
	n := len(candles)
	if n < minLen {
		return nil
	}

	lower, upper := utils.LbKeoghEnvelope(seedVec, resampledLength)

	var matches []match
	var mu sync.Mutex

	for windowLen := minLen; windowLen <= maxLen && windowLen <= n; windowLen++ {
		vecs := make([][]float64, n-windowLen+1)
		for i := 0; i+windowLen <= n; i++ {
			vecs[i] = getPricesVec(candles[i:i+windowLen], resampledLength)
		}

		var wg sync.WaitGroup
		matchesCh := make(chan match, n)
		tasks := make(chan int, n-windowLen+1)

		for i := 0; i < runtime.NumCPU(); i++ {
			wg.Add(1)
			go s.matchWorker(tasks, &wg, matchesCh, seedVec, lower, upper, vecs,
				windowLen, ticker, candles, tolerance, resampledLength)
		}

		for winStart := 0; winStart <= n-windowLen; winStart++ {
			tasks <- winStart
		}

		close(tasks)
		wg.Wait()
		close(matchesCh)

		for m := range matchesCh {
			mu.Lock()
			matches = append(matches, m)
			mu.Unlock()
		}
	}

	return matches
}

// matchWorker обрабатывает задачи поиска совпадений
func (s *Scanner) matchWorker(tasks <-chan int, wg *sync.WaitGroup, matchesCh chan<- match,
	seedVec, lower, upper []float64, cacheVecs [][]float64, windowLen int,
	ticker string, candles []models.Candle, tolerance float64, resampledLength int) {
	defer wg.Done()

	maxCost := tolerance * float64(resampledLength)

	for winStart := range tasks {
		if winStart < 0 || winStart >= len(cacheVecs) {
			continue
		}
		candlesVec := cacheVecs[winStart]

		lb := utils.LbKeoghDistance(seedVec, lower, upper, candlesVec)
		if lb > maxCost {
			continue
		}

		d := utils.DTW(seedVec, candlesVec, maxCost)
		if d < 0 || d > maxCost {
			continue
		}

		endIdx := winStart + windowLen - 1
		if endIdx >= len(candles) {
			continue
		}

		matchCandles := candles[winStart : endIdx+1]
		if len(matchCandles) == 0 {
			continue
		}

		seg := models.ChartSegment{
			Ticker:  ticker,
			From:    matchCandles[0].Date,
			To:      matchCandles[len(matchCandles)-1].Date,
			Candles: matchCandles,
		}

		normalizedDistance := d / float64(resampledLength)

		matchesCh <- match{
			Segment:  seg,
			Distance: normalizedDistance,
		}
	}
}

// removeOverlaps удаляет наложенные сегменты, оставляя лучшие по DTW расстоянию
func removeOverlaps(matches []match) []match {
	if len(matches) == 0 {
		return matches
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Distance < matches[j].Distance
	})

	var result []match
	for _, m := range matches {
		overlaps := false
		for _, existing := range result {
			if isOverlap(m.Segment, existing.Segment) {
				overlaps = true
				break
			}
		}
		if !overlaps {
			result = append(result, m)
		}
	}

	return result
}

// isOverlap проверяет, накладываются ли два сегмента друг на друга
func isOverlap(seg1, seg2 models.ChartSegment) bool {
	if seg1.Ticker != seg2.Ticker {
		return false
	}

	return !(seg1.To.Before(seg2.From) || seg1.To.Equal(seg2.From) ||
		seg2.To.Before(seg1.From) || seg2.To.Equal(seg1.From))
}

// getPricesVec извлекает цены закрытия, нормализует и ресемплирует их
func getPricesVec(candles []models.Candle, resampledLength int) []float64 {
	prices := make([]float64, len(candles))
	for i := range candles {
		prices[i] = candles[i].Close
	}
	normSeed := utils.ZNormalize(prices)
	vec := utils.Resample(normSeed, resampledLength)
	return vec
}
