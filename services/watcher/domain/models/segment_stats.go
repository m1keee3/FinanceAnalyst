package models

type SegmentStats struct {
	Segment      *ChartSegment
	TotalMatches int
	PriceChange  float64
	Probability  float64
}
