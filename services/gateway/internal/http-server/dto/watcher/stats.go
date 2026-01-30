package watcher

import "time"

type StatsResponse struct {
	Items []Stats `json:"items"`
}

type Stats struct {
	Segment      *Segment `json:"segment"`
	PatternType  string   `json:"pattern_type"`
	TotalMatches int      `json:"total_matches"`
	PriceChange  float64  `json:"price_change"`
	Probability  float64  `json:"probability"`
}

type Segment struct {
	Ticker string    `json:"ticker"`
	From   time.Time `json:"from"`
	To     time.Time `json:"to"`
}
