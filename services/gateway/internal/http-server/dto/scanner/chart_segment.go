package scanner

import "time"

type ChartSegment struct {
	Ticker string    `json:"ticker"`
	From   time.Time `json:"from"`
	To     time.Time `json:"to"`
}
