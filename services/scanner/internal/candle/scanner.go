package candle

import (
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
// tailLen — длина начального хвоста в свечах, tolerance — допуск по проценто-изменению для основной части,
// searchFrom/searchTo — период, в котором искать по каждому тикеру.
func (s *Scanner) FindMatches(segment models.ChartSegment, tickers []string, searchFrom, searchTo time.Time, options *ScanOptions) ([]models.ChartSegment, error) {
	return nil, nil
}

// ComputeStats считает статистику по совпадениям для заданного сегмента.
func (s *Scanner) ComputeStats(matches []models.ChartSegment, horizonDays int) (*models.ScanStats, error) {
	return nil, nil
}
