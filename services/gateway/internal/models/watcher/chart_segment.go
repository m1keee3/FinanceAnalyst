package watcher

import "time"

type ChartSegment struct {
	Ticker string
	From   time.Time
	To     time.Time
}
