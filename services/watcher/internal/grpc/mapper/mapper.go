package mapper

import (
	"github.com/m1keee3/FinanceAnalyst/services/watcher/domain/models"
	watcherv1 "github.com/m1keee3/FinanceAnalyst/services/watcher/proto-gen/watcher/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToGetStatsResponse(stats []*models.SegmentStats) *watcherv1.GetStatsResponse {
	protoStats := make([]*watcherv1.SegmentStats, len(stats))

	for i, s := range stats {
		protoStats[i] = toProtoStats(s)
	}

	return &watcherv1.GetStatsResponse{
		Stats: protoStats,
	}
}

func toProtoStats(stats *models.SegmentStats) *watcherv1.SegmentStats {
	return &watcherv1.SegmentStats{
		Segment:      toProtoChartSegment(stats.Segment),
		TotalMatches: int32(stats.TotalMatches),
		PriceChange:  stats.PriceChange,
		Probability:  stats.Probability,
		PatternType:  string(stats.PatternType),
	}
}

func toProtoChartSegment(seg *models.ChartSegment) *watcherv1.Segment {
	return &watcherv1.Segment{
		Ticker: seg.Ticker,
		From:   timestamppb.New(seg.From),
		To:     timestamppb.New(seg.To),
	}
}
