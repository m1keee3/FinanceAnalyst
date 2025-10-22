package chart

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
func NewScanQuery(req *scannerv1.ChartScanRequest) *ScanQuery {
	segment := mapper.FromProtoChartSegment(req.GetSegment())
	options := FromProtoChartScanOptions(req.GetOptions())

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

// FromProtoChartScanOptions конвертирует proto ChartScanOptions в ScanOptions
func FromProtoChartScanOptions(proto *scannerv1.ChartScanOptions) ScanOptions {
	if proto == nil {
		return ScanOptions{}
	}

	return ScanOptions{
		MinScale:  proto.GetMinScale(),
		MaxScale:  proto.GetMaxScale(),
		Tolerance: proto.GetTolerance(),
	}
}
