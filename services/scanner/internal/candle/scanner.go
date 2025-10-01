package candle

import (
	"log"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/m1keee3/FinanceAnalyst/common/models"
)

type Fetcher interface {
	Fetch(ticker string, from, to time.Time) ([]models.Candle, error)
}

type Scanner struct {
	fetcher Fetcher
}

func NewScanner(fetcher Fetcher) *Scanner {
	return &Scanner{fetcher: fetcher}
}

// ScanOptions определяет параметры сравнения свечей
type ScanOptions struct {
	TailLen         int
	BodyTolerance   float64
	ShadowTolerance float64
}

func (o *ScanOptions) withDefaults() ScanOptions {
	out := ScanOptions{TailLen: 0, BodyTolerance: 0.1, ShadowTolerance: 0.1}
	if o == nil {
		return out
	}
	if o.TailLen > 0 {
		out.TailLen = o.TailLen
	}
	if o.BodyTolerance > 0 {
		out.BodyTolerance = o.BodyTolerance
	}
	if o.ShadowTolerance > 0 {
		out.ShadowTolerance = o.ShadowTolerance
	}
	return out
}

// FindMatches ищет совпадения для заданного сегмента на указанных тикерах по всему периоду поиска.
// tailLen — длина начального хвоста в свечах, tolerance — допуск по процентно-изменению для основной части,
// searchFrom/searchTo — период, в котором искать по каждому тикеру.
func (s *Scanner) FindMatches(segment models.ChartSegment, tickers []string, searchFrom, searchTo time.Time, options *ScanOptions) ([]models.ChartSegment, error) {
	if s == nil || s.fetcher == nil {
		return nil, nil
	}

	if len(segment.Candles) == 0 || len(tickers) == 0 {
		return nil, nil
	}

	opts := options.withDefaults()
	L := len(segment.Candles)
	if opts.TailLen < 0 {
		opts.TailLen = 0
	}
	if opts.TailLen > L {
		opts.TailLen = L
	}

	normSegment := models.NormalizeCandles(segment.Candles)

	targetTailSign := tailSign(normSegment[:opts.TailLen])

	// Параллельная обработка тикеров
	workerCount := runtime.NumCPU()
	if workerCount < 2 {
		workerCount = 2
	}

	tickerCh := make(chan string)
	matchCh := make(chan models.ChartSegment, 1024)
	errCh := make(chan error, workerCount)
	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for ticker := range tickerCh {
			candles, err := s.fetcher.Fetch(ticker, searchFrom, searchTo)
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
				continue
			}
			for i := 0; i+L <= len(candles); i++ {
				window := candles[i : i+L]
				normWindow := models.NormalizeCandles(window)
				if opts.TailLen > 0 {
					if tailSign(normWindow[:opts.TailLen]) != targetTailSign {
						continue
					}
				}
				if similarCoreWithShadows(
					normWindow[opts.TailLen:],
					normSegment[opts.TailLen:],
					opts.BodyTolerance,
					opts.ShadowTolerance,
				) {
					match := models.ChartSegment{
						Ticker:  ticker,
						From:    window[0].Date,
						To:      window[len(window)-1].Date,
						Candles: append([]models.Candle(nil), window...),
					}
					if !IsOverlap(segment, match) {
						matchCh <- match
					}
				}
			}
		}
	}

	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go worker()
	}

	go func() {
		for _, t := range tickers {
			tickerCh <- t
		}
		close(tickerCh)
	}()

	var matches []models.ChartSegment
	done := make(chan struct{})
	go func() {
		for m := range matchCh {
			matches = append(matches, m)
		}
		close(done)
	}()

	wg.Wait()
	close(matchCh)
	<-done

	for {
		select {
		case e := <-errCh:
			log.Printf("error in worker: %v", e)
		default:
			return matches, nil
		}
	}
}

// tailSign возвращает знак суммарного движения свечей (по цене Close-Open)
func tailSign(candles []models.Candle) bool {
	if len(candles) == 0 {
		return true
	}
	return math.Signbit(candles[0].Open - candles[len(candles)-1].Close)
}

// similarCoreWithShadows сравнивает основную часть сегмента по телу и теням с допусками
func similarCoreWithShadows(window []models.Candle, targetCandles []models.Candle, bodyTolerance, shadowTolerance float64) bool {
	if len(window) == 0 || len(targetCandles) == 0 {
		return false
	}

	for i := 0; i < len(window); i++ { // обработка свечей в паттерне
		winSign := math.Signbit(window[i].Open - window[i].Close)
		targetSign := math.Signbit(targetCandles[i].Open - targetCandles[i].Close)

		if winSign != targetSign {
			return false // если свечи в разных направлениях, паттерн не совпадает
		}

		// Верх свечи open
		if math.Abs(window[i].Open-targetCandles[i].Open) > bodyTolerance {
			return false
		}

		// Низ свечи close
		if math.Abs(window[i].Close-targetCandles[i].Close) > bodyTolerance {
			return false
		}

		// Верхняя тень (high - max(open,close))
		candleUpper := window[i].High - math.Max(window[i].Open, window[i].Close)
		patternUpper := targetCandles[i].High - math.Max(
			targetCandles[i].Open,
			targetCandles[i].Close)
		if math.Abs(candleUpper-patternUpper) > shadowTolerance {
			return false
		}

		// Нижняя тень (min(open,close) - low)
		candleLower := math.Min(window[i].Open, window[i].Close) - window[i].Low
		patternLower := math.Min(
			targetCandles[i].Open,
			targetCandles[i].Close) - targetCandles[i].Low
		if math.Abs(candleLower-patternLower) > shadowTolerance {
			return false
		}
	}

	return true
}

// IsOverlap проверяет, накладываются ли два сегмента друг на друга.
// Сегменты считаются наложенными, если они относятся к одному тикеру
// и их временные интервалы пересекаются.
func IsOverlap(seg1 models.ChartSegment, seg2 models.ChartSegment) bool {
	if seg1.Ticker != seg2.Ticker {
		return false
	}

	return !(seg1.To.Before(seg2.From) || seg1.To.Equal(seg2.From) ||
		seg2.To.Before(seg1.From) || seg2.To.Equal(seg1.From))
}
