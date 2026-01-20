package mapper

import (
	"github.com/m1keee3/FinanceAnalyst/services/watcher/domain/models"
	"github.com/m1keee3/FinanceAnalyst/services/watcher/internal/services/watcher/models/candle"
	"github.com/m1keee3/FinanceAnalyst/services/watcher/internal/services/watcher/models/chart"
	scannerv1 "github.com/m1keee3/FinanceAnalyst/services/watcher/proto-gen/scanner/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToCandleStatsRequest(q *candle.ScanQuery) *scannerv1.CandleStatsRequest {
	return &scannerv1.CandleStatsRequest{
		Scan:        toCandleScanRequest(q),
		DaysToWatch: int32(q.Options.DaysToWatch),
	}
}

func ToChartStatsRequest(q *chart.ScanQuery) *scannerv1.ChartStatsRequest {
	return &scannerv1.ChartStatsRequest{
		Scan:        toChartScanRequest(q),
		DaysToWatch: int32(q.Options.DaysToWatch),
	}
}

func FromCandleStatsResponse(resp *scannerv1.CandleStatsResponse) *models.SegmentStats {
	return &models.SegmentStats{
		TotalMatches: int(resp.Stats.TotalMatches),
		PriceChange:  resp.Stats.PriceChange,
		Probability:  resp.Stats.Probability,
	}
}

func FromChartStatsResponse(resp *scannerv1.ChartStatsResponse) *models.SegmentStats {
	return &models.SegmentStats{
		TotalMatches: int(resp.Stats.TotalMatches),
		PriceChange:  resp.Stats.PriceChange,
		Probability:  resp.Stats.Probability,
	}
}

func toCandleScanRequest(q *candle.ScanQuery) *scannerv1.CandleScanRequest {
	return &scannerv1.CandleScanRequest{
		Segment:    toProtoChartSegment(q.Segment),
		Options:    toProtoCandleScanOptions(q.Options),
		SearchFrom: timestamppb.New(q.SearchFrom),
		SearchTo:   timestamppb.New(q.SearchTo),
		Tickers:    q.Tickers,
	}
}

func toChartScanRequest(q *chart.ScanQuery) *scannerv1.ChartScanRequest {
	return &scannerv1.ChartScanRequest{
		Segment:    toProtoChartSegment(q.Segment),
		Options:    toProtoChartScanOptions(q.Options),
		SearchFrom: timestamppb.New(q.SearchFrom),
		SearchTo:   timestamppb.New(q.SearchTo),
		Tickers:    q.Tickers,
	}
}

func toProtoChartSegment(segment *models.ChartSegment) *scannerv1.ChartSegment {
	return &scannerv1.ChartSegment{
		Ticker: segment.Ticker,
		From:   timestamppb.New(segment.From),
		To:     timestamppb.New(segment.To),
	}
}

func toProtoCandleScanOptions(opts *candle.ScanOptions) *scannerv1.CandleScanOptions {
	return &scannerv1.CandleScanOptions{
		TailLen:         int32(opts.TailLen),
		ShadowTolerance: opts.ShadowTolerance,
		BodyTolerance:   opts.BodyTolerance,
	}
}

func toProtoChartScanOptions(options *chart.ScanOptions) *scannerv1.ChartScanOptions {
	return &scannerv1.ChartScanOptions{
		MinScale:  options.MinScale,
		MaxScale:  options.MaxScale,
		Tolerance: options.Tolerance,
	}
}
