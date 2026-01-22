package mapper

import (
	"github.com/m1keee3/FinanceAnalyst/services/scanner/domain/models"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/models/candle"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/services/models/chart"
	scannerv1 "github.com/m1keee3/FinanceAnalyst/services/scanner/proto-gen/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MatchesToCandleMatchResponse(matches []models.ChartSegment) *scannerv1.CandleMatchesResponse {
	protoMatches := make([]*scannerv1.ChartSegment, len(matches))

	for i, m := range matches {
		protoMatches[i] = toProtoChartSegment(m)
	}

	return &scannerv1.CandleMatchesResponse{
		Matches: protoMatches,
	}
}

func MatchesToChartMatchResponse(matches []models.ChartSegment) *scannerv1.ChartMatchesResponse {
	protoMatches := make([]*scannerv1.ChartSegment, len(matches))

	for i, m := range matches {
		protoMatches[i] = toProtoChartSegment(m)
	}

	return &scannerv1.ChartMatchesResponse{
		Matches: protoMatches,
	}
}

func ScanStatsToCandleStatsResponse(stats *models.ScanStats) *scannerv1.CandleStatsResponse {
	return &scannerv1.CandleStatsResponse{
		Stats: &scannerv1.ScanStats{
			TotalMatches: int32(stats.TotalMatches),
			PriceChange:  stats.PriceChange,
			Probability:  stats.Probability,
		},
	}
}

func ScanStatsToChartStatsResponse(stats *models.ScanStats) *scannerv1.ChartStatsResponse {
	return &scannerv1.ChartStatsResponse{
		Stats: &scannerv1.ScanStats{
			TotalMatches: int32(stats.TotalMatches),
			PriceChange:  stats.PriceChange,
			Probability:  stats.Probability,
		},
	}
}

func CandleScanRequestToScanQuery(req *scannerv1.CandleScanRequest) *candle.ScanQuery {
	segment := fromProtoChartSegment(req.GetSegment())
	options := fromProtoCandleScanOptions(req.GetOptions())

	return &candle.ScanQuery{
		Segment:    segment,
		Options:    options,
		SearchFrom: req.GetSearchFrom().AsTime(),
		SearchTo:   req.GetSearchTo().AsTime(),
		Tickers:    req.GetTickers(),
	}
}

func ChartScanRequestToScanQuery(req *scannerv1.ChartScanRequest) *chart.ScanQuery {
	segment := fromProtoChartSegment(req.GetSegment())
	options := fromProtoChartScanOptions(req.GetOptions())

	return &chart.ScanQuery{
		Segment:    segment,
		Options:    options,
		SearchFrom: req.GetSearchFrom().AsTime(),
		SearchTo:   req.GetSearchTo().AsTime(),
		Tickers:    req.GetTickers(),
	}
}

func CandleStatsRequestToStatsQuery(req *scannerv1.CandleStatsRequest) *candle.StatsQuery {
	return &candle.StatsQuery{
		ScanQuery:   CandleScanRequestToScanQuery(req.GetScan()),
		DaysToWatch: int(req.GetDaysToWatch()),
	}
}

func ChartStatsRequestToStatsQuery(req *scannerv1.ChartStatsRequest) *chart.StatsQuery {
	return &chart.StatsQuery{
		ScanQuery:   ChartScanRequestToScanQuery(req.GetScan()),
		DaysToWatch: int(req.GetDaysToWatch()),
	}
}

func fromProtoChartSegment(proto *scannerv1.ChartSegment) models.ChartSegment {

	return models.ChartSegment{
		Ticker: proto.Ticker,
		From:   proto.GetFrom().AsTime(),
		To:     proto.GetTo().AsTime(),
	}
}

func toProtoChartSegment(segment models.ChartSegment) *scannerv1.ChartSegment {

	return &scannerv1.ChartSegment{
		Ticker: segment.Ticker,
		From:   timestamppb.New(segment.From),
		To:     timestamppb.New(segment.To),
	}
}

func fromProtoCandleScanOptions(proto *scannerv1.CandleScanOptions) candle.ScanOptions {
	if proto == nil {
		return candle.ScanOptions{}
	}

	return candle.ScanOptions{
		TailLen:         int(proto.GetTailLen()),
		BodyTolerance:   proto.GetBodyTolerance(),
		ShadowTolerance: proto.GetShadowTolerance(),
	}
}

func fromProtoChartScanOptions(proto *scannerv1.ChartScanOptions) chart.ScanOptions {
	if proto == nil {
		return chart.ScanOptions{}
	}

	return chart.ScanOptions{
		MinScale:  proto.GetMinScale(),
		MaxScale:  proto.GetMaxScale(),
		Tolerance: proto.GetTolerance(),
	}
}
