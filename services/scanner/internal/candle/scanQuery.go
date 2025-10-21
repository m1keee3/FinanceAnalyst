package candle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/m1keee3/FinanceAnalyst/services/scanner/domain/models"
	scannerv1 "github.com/m1keee3/FinanceAnalyst/services/scanner/proto-gen/v1"
)

type ScanQuery struct {
	Segment    models.ChartSegment
	Options    ScanOptions
	SearchFrom time.Time
	SearchTo   time.Time
	Tickers    []string
}

// NewScanQuery создает ScanQuery из proto запроса
func NewScanQuery(req *scannerv1.CandleScanRequest) *ScanQuery {
	segment := protoToChartSegment(req.GetSegment())
	options := protoToCandleScanOptions(req.GetOptions())

	return &ScanQuery{
		Segment:    segment,
		Options:    options,
		SearchFrom: req.GetSearchFrom().AsTime(),
		SearchTo:   req.GetSearchTo().AsTime(),
		Tickers:    req.GetTickers(),
	}
}

func (q ScanQuery) Hash() string {
	h := sha256.New()
	enc := json.NewEncoder(h)
	_ = enc.Encode(q.Segment.Candles)
	_ = enc.Encode(q.Options)
	_ = enc.Encode(q.SearchFrom.Unix())
	_ = enc.Encode(q.SearchTo.Unix())
	_ = enc.Encode(q.Tickers)
	return hex.EncodeToString(h.Sum(nil))
}

// protoToChartSegment конвертирует proto ChartSegment в models.ChartSegment
func protoToChartSegment(proto *scannerv1.ChartSegment) models.ChartSegment {
	if proto == nil {
		return models.ChartSegment{}
	}

	candles := make([]models.Candle, len(proto.GetCandles()))
	for i, c := range proto.GetCandles() {
		candles[i] = models.Candle{
			Date:  c.GetDate().AsTime(),
			Open:  c.GetOpen(),
			High:  c.GetHigh(),
			Low:   c.GetLow(),
			Close: c.GetClose(),
		}
	}

	return models.ChartSegment{
		Ticker:  proto.GetTicker(),
		From:    proto.GetFrom().AsTime(),
		To:      proto.GetTo().AsTime(),
		Candles: candles,
	}
}

// protoToCandleScanOptions конвертирует proto CandleScanOptions в ScanOptions
func protoToCandleScanOptions(proto *scannerv1.CandleScanOptions) ScanOptions {
	if proto == nil {
		return ScanOptions{}
	}

	return ScanOptions{
		TailLen:         int(proto.GetTailLen()),
		BodyTolerance:   proto.GetBodyTolerance(),
		ShadowTolerance: proto.GetShadowTolerance(),
	}
}
