package mapper

import (
	scannerDTO "github.com/m1keee3/FinanceAnalyst/services/gateway/internal/http-server/dto/scanner"
	candleDTO "github.com/m1keee3/FinanceAnalyst/services/gateway/internal/http-server/dto/scanner/candle"
	chartDTO "github.com/m1keee3/FinanceAnalyst/services/gateway/internal/http-server/dto/scanner/chart"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/models/scanner"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/models/scanner/candle"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/models/scanner/chart"
)

func ToCandleStatsRequest(req *candleDTO.StatsRequest) *candle.StatsRequest {
	return &candle.StatsRequest{
		ScanRequest: ToCandleScanRequest(req.ScanRequest),
		DaysToWatch: req.DaysToWatch,
	}
}

func ToChartStatsRequest(req *chartDTO.StatsRequest) *chart.StatsRequest {
	return &chart.StatsRequest{
		ScanRequest: ToChartScanRequest(req.ScanRequest),
		DaysToWatch: req.DaysToWatch,
	}
}

func ToCandleScanRequest(req *candleDTO.ScanRequest) *candle.ScanRequest {
	return &candle.ScanRequest{
		Segment:    toChartSegment(req.Segment),
		Options:    toCandleScanOptions(req.Options),
		SearchFrom: req.SearchFrom,
		SearchTo:   req.SearchTo,
		Tickers:    req.Tickers,
	}
}

func ToChartScanRequest(req *chartDTO.ScanRequest) *chart.ScanRequest {
	return &chart.ScanRequest{
		Segment:    toChartSegment(req.Segment),
		Options:    toChartScanOptions(req.Options),
		SearchFrom: req.SearchFrom,
		SearchTo:   req.SearchTo,
		Tickers:    req.Tickers,
	}
}

func FromMatches(matches []*scanner.ChartSegment) []*scannerDTO.ChartSegment {
	res := make([]*scannerDTO.ChartSegment, len(matches))

	for i, m := range matches {
		res[i] = fromChartSegment(m)
	}

	return res
}

func FromScanStats(stats *scanner.ScanStats) scannerDTO.ScanStats {
	return scannerDTO.ScanStats{
		TotalMatches: stats.TotalMatches,
		PriceChange:  stats.PriceChange,
		Probability:  stats.Probability,
	}
}

func toCandleScanOptions(opts *candleDTO.ScanOptions) *candle.ScanOptions {
	return &candle.ScanOptions{
		TailLen:         opts.TailLen,
		BodyTolerance:   opts.BodyTolerance,
		ShadowTolerance: opts.ShadowTolerance,
	}
}

func toChartScanOptions(opts *chartDTO.ScanOptions) *chart.ScanOptions {
	return &chart.ScanOptions{
		MinScale:  opts.MinScale,
		MaxScale:  opts.MaxScale,
		Tolerance: opts.Tolerance,
	}
}

func toChartSegment(seg *scannerDTO.ChartSegment) *scanner.ChartSegment {
	return &scanner.ChartSegment{
		Ticker: seg.Ticker,
		From:   seg.From,
		To:     seg.To,
	}
}

func fromChartSegment(seg *scanner.ChartSegment) *scannerDTO.ChartSegment {
	return &scannerDTO.ChartSegment{
		Ticker: seg.Ticker,
		From:   seg.From,
		To:     seg.To,
	}
}
