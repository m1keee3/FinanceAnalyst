package candle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/m1keee3/FinanceAnalyst/services/scanner/domain/models"
	"github.com/m1keee3/FinanceAnalyst/services/scanner/internal/mapper"
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
	segment := mapper.FromProtoChartSegment(req.GetSegment())
	options := FromProtoCandleScanOptions(req.GetOptions())

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

// FromProtoCandleScanOptions конвертирует proto CandleScanOptions в ScanOptions
func FromProtoCandleScanOptions(proto *scannerv1.CandleScanOptions) ScanOptions {
	if proto == nil {
		return ScanOptions{}
	}

	return ScanOptions{
		TailLen:         int(proto.GetTailLen()),
		BodyTolerance:   proto.GetBodyTolerance(),
		ShadowTolerance: proto.GetShadowTolerance(),
	}
}
