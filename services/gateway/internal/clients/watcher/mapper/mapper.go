package mapper

import (
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/models/watcher"
	watcherv1 "github.com/m1keee3/FinanceAnalyst/services/gateway/proto-gen/watcher/v1"
)

func FromStatsResponse(resp *watcherv1.GetStatsResponse) []*watcher.SegmentStats {
	res := make([]*watcher.SegmentStats, len(resp.Stats))

	for i, s := range resp.Stats {
		res[i] = fromProtoStats(s)
	}

	return res
}

func fromProtoStats(s *watcherv1.SegmentStats) *watcher.SegmentStats {
	return &watcher.SegmentStats{
		Segment:      fromProtoSegment(s.Segment),
		TotalMatches: int(s.TotalMatches),
		PriceChange:  s.PriceChange,
		Probability:  s.Probability,
		PatternType:  watcher.PatternType(s.PatternType),
	}
}

func fromProtoSegment(stats *watcherv1.Segment) *watcher.ChartSegment {
	return &watcher.ChartSegment{
		Ticker: stats.Ticker,
		From:   stats.From.AsTime(),
		To:     stats.To.AsTime(),
	}
}
