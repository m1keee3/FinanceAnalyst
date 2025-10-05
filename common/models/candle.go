package models

import (
	"cmp"
	"slices"
	"time"
)

type Candle struct {
	Date  time.Time
	Open  float64
	High  float64
	Low   float64
	Close float64
}

func (c Candle) Normalize(min, max float64) Candle {
	rangeVal := max - min
	if rangeVal == 0 {
		rangeVal = 1
	}

	c.Open = (c.Open - min) / rangeVal
	c.High = (c.High - min) / rangeVal
	c.Low = (c.Low - min) / rangeVal
	c.Close = (c.Close - min) / rangeVal

	return c
}

func NormalizeCandles(candles []Candle) []Candle {
	if len(candles) == 0 {
		return nil
	}

	res := make([]Candle, 0, len(candles))

	maxHigh := slices.MaxFunc(
		candles, func(a, b Candle) int {
			return cmp.Compare(a.High, b.High)
		}).High
	minLow := slices.MinFunc(
		candles, func(a, b Candle) int {
			return cmp.Compare(a.Low, b.Low)
		}).Low

	for i := range candles {
		res = append(res, candles[i].Normalize(minLow, maxHigh))
	}
	return res
}
