package models

type PatternType string

const (
	PatternTypeCandle PatternType = "candle"
	PatternTypeChart  PatternType = "chart"
)

type SegmentStats struct {
	Segment      *ChartSegment
	TotalMatches int
	PriceChange  float64
	Probability  float64
	PatternType  PatternType
}
