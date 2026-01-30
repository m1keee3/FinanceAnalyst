package mapper

import (
	watcherDTO "github.com/m1keee3/FinanceAnalyst/services/gateway/internal/http-server/dto/watcher"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/models/watcher"
)

func FromStats(stats []*watcher.SegmentStats) []*watcherDTO.Stats {
	res := make([]*watcherDTO.Stats, len(stats))

	for i, s := range stats {
		res[i] = fromStats(s)
	}

	return res
}

func fromStats(stats *watcher.SegmentStats) *watcherDTO.Stats {
	return &watcherDTO.Stats{
		Segment:      fromSegment(stats.Segment),
		PatternType:  string(stats.PatternType),
		TotalMatches: stats.TotalMatches,
		PriceChange:  stats.PriceChange,
		Probability:  stats.Probability,
	}
}

func fromSegment(seg *watcher.ChartSegment) *watcherDTO.Segment {
	return &watcherDTO.Segment{
		Ticker: seg.Ticker,
		From:   seg.From,
		To:     seg.To,
	}
}
