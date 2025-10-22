package stats

import (
	"fmt"
	"log"
	"time"

	"github.com/m1keee3/FinanceAnalyst/common/models"
)

type Fetcher interface {
	Fetch(ticker string, from, to time.Time) ([]models.Candle, error)
}

type Evaluator struct {
	fetcher Fetcher
}

func NewEvaluator(fetcher Fetcher) *Evaluator {
	return &Evaluator{fetcher: fetcher}
}

// ComputeStats считает статистику по совпадениям для заданного сегмента.
// daysToWatch это количество свечей после сегмента, которые надо рассмотреть, если daysToWatch = 0, то алгоритм рассматривает свечи пока они идут в одном направлении
func (e *Evaluator) ComputeStats(matches []models.ChartSegment, daysToWatch int) (*models.ScanStats, error) {
	if e == nil || e.fetcher == nil {
		return &models.ScanStats{}, nil
	}

	if len(matches) == 0 {
		return &models.ScanStats{TotalMatches: 0, PriceChange: 0, Probability: 0}, nil
	}

	if daysToWatch == 0 {
		return e.computeLineStats(matches)
	}

	var considered int
	var posCtr int
	var posSumChange float64
	var negSumChange float64

	for _, m := range matches {
		candles, err := e.fetcher.Fetch(m.Ticker, m.To.AddDate(0, 0, 1), m.To.AddDate(0, 0, daysToWatch))
		if err != nil {
			log.Print(fmt.Errorf("error fetching candles for %s: %w", m.Ticker, err))
			continue
		}
		for i := 1; len(candles) < daysToWatch && i < 2; i++ {
			candles, err = e.fetcher.Fetch(m.Ticker, m.To.AddDate(0, 0, 1), m.To.AddDate(0, 0, i+daysToWatch))
			if err != nil {
				log.Print(fmt.Errorf("error fetching candles for %s: %w", m.Ticker, err))
				continue
			}
		}

		if len(candles) == 0 {
			continue
		}

		var delta float64

		limit := daysToWatch
		if limit > len(candles) {
			limit = len(candles)
		}
		for j := 0; j < limit; j++ {
			delta += candles[j].Close - candles[j].Open
		}

		considered++
		if delta >= 0 {
			posCtr++
			posSumChange += delta / candles[0].Open
		} else {
			delta = -delta
			negSumChange -= delta / candles[0].Open
		}

	}

	if considered == 0 {
		return &models.ScanStats{TotalMatches: 0, PriceChange: 0, Probability: 0}, nil
	}

	var avgChange float64
	var prob float64
	if posCtr > considered-posCtr {
		avgChange = posSumChange / float64(posCtr)
		prob = float64(posCtr) / float64(considered)
	} else {
		avgChange = negSumChange / float64(considered-posCtr)
		prob = float64(considered-posCtr) / float64(considered)
	}

	return &models.ScanStats{
		TotalMatches: considered,
		PriceChange:  avgChange,
		Probability:  prob,
	}, nil
}

func (s *Evaluator) computeLineStats(matches []models.ChartSegment) (*models.ScanStats, error) {
	var considered int
	var posCtr int
	var posSumChange float64
	var negSumChange float64

	for _, m := range matches {
		candles, err := s.fetcher.Fetch(m.Ticker, m.To.AddDate(0, 0, 1), m.To.AddDate(0, 0, 30))
		if err != nil {
			log.Print(fmt.Errorf("error fetching candles for %s: %w", m.Ticker, err))
			continue
		}

		if len(candles) == 0 {
			continue
		}

		var delta float64

		sign := candles[0].Close-candles[0].Open >= 0
		for _, c := range candles {
			dif := c.Close - c.Open
			if dif >= 0 != sign {
				break
			}
			delta += dif
		}

		considered++
		if delta >= 0 {
			posCtr++
			posSumChange += delta / candles[0].Open
		} else {
			delta = -delta
			negSumChange -= delta / candles[0].Open
		}

	}

	if considered == 0 {
		return &models.ScanStats{TotalMatches: 0, PriceChange: 0, Probability: 0}, nil
	}

	var avgChange float64
	var prob float64
	if posCtr > considered-posCtr {
		avgChange = posSumChange / float64(posCtr)
		prob = float64(posCtr) / float64(considered)
	} else {
		avgChange = negSumChange / float64(considered-posCtr)
		prob = float64(considered-posCtr) / float64(considered)
	}

	return &models.ScanStats{
		TotalMatches: considered,
		PriceChange:  avgChange,
		Probability:  prob,
	}, nil
}
