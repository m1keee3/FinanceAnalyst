package mapper

import (
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/models/scanner"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/models/scanner/candle"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/models/scanner/chart"
	scannerv1 "github.com/m1keee3/FinanceAnalyst/services/gateway/proto-gen/scanner/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToCandleScanRequest(q *candle.ScanRequest) *scannerv1.CandleScanRequest {
	return &scannerv1.CandleScanRequest{
		Segment:    toProtoSegment(q.Segment),
		Options:    toProtoCandleOptions(q.Options),
		SearchFrom: timestamppb.New(q.SearchFrom),
		SearchTo:   timestamppb.New(q.SearchTo),
		Tickers:    q.Tickers,
	}
}

func ToChartScanRequest(q *chart.ScanRequest) *scannerv1.ChartScanRequest {
	return &scannerv1.ChartScanRequest{
		Segment:    toProtoSegment(q.Segment),
		Options:    toProtoChartOptions(q.Options),
		SearchFrom: timestamppb.New(q.SearchFrom),
		SearchTo:   timestamppb.New(q.SearchTo),
		Tickers:    q.Tickers,
	}
}

func ToCandleStatsRequest(q *candle.StatsRequest) *scannerv1.CandleStatsRequest {
	return &scannerv1.CandleStatsRequest{
		Scan:        ToCandleScanRequest(q.ScanRequest),
		DaysToWatch: int32(q.DaysToWatch),
	}
}

func ToChartStatsRequest(q *chart.StatsRequest) *scannerv1.ChartStatsRequest {
	return &scannerv1.ChartStatsRequest{
		Scan:        ToChartScanRequest(q.ScanRequest),
		DaysToWatch: int32(q.DaysToWatch),
	}
}

func FromCandleMatchesResponse(resp *scannerv1.CandleMatchesResponse) []*scanner.ChartSegment {
	res := make([]*scanner.ChartSegment, len(resp.Matches))

	for i, m := range resp.Matches {
		res[i] = fromProtoSegment(m)
	}

	return res
}

func FromChartMatchesResponse(resp *scannerv1.ChartMatchesResponse) []*scanner.ChartSegment {
	res := make([]*scanner.ChartSegment, len(resp.Matches))

	for i, m := range resp.Matches {
		res[i] = fromProtoSegment(m)
	}

	return res
}

func FromCandleStatsResponse(resp *scannerv1.CandleStatsResponse) *scanner.ScanStats {
	return &scanner.ScanStats{
		TotalMatches: int(resp.Stats.TotalMatches),
		PriceChange:  resp.Stats.PriceChange,
		Probability:  resp.Stats.Probability,
	}
}

func FromChartStatsResponse(resp *scannerv1.ChartStatsResponse) *scanner.ScanStats {
	return &scanner.ScanStats{
		TotalMatches: int(resp.Stats.TotalMatches),
		PriceChange:  resp.Stats.PriceChange,
		Probability:  resp.Stats.Probability,
	}
}

func toProtoSegment(segment *scanner.ChartSegment) *scannerv1.ChartSegment {
	return &scannerv1.ChartSegment{
		Ticker: segment.Ticker,
		From:   timestamppb.New(segment.From),
		To:     timestamppb.New(segment.To),
	}
}

func fromProtoSegment(segment *scannerv1.ChartSegment) *scanner.ChartSegment {
	return &scanner.ChartSegment{
		Ticker: segment.Ticker,
		From:   segment.From.AsTime(),
		To:     segment.To.AsTime(),
	}
}

func toProtoCandleOptions(opts *candle.ScanOptions) *scannerv1.CandleScanOptions {
	return &scannerv1.CandleScanOptions{
		TailLen:         int32(opts.TailLen),
		BodyTolerance:   opts.BodyTolerance,
		ShadowTolerance: opts.ShadowTolerance,
	}
}

func toProtoChartOptions(opts *chart.ScanOptions) *scannerv1.ChartScanOptions {
	return &scannerv1.ChartScanOptions{
		MinScale:  opts.MinScale,
		MaxScale:  opts.MaxScale,
		Tolerance: opts.Tolerance,
	}
}
