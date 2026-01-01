package candle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/m1keee3/FinanceAnalyst/services/scanner/domain/models"
)

type ScanQuery struct {
	Segment    models.ChartSegment
	Options    ScanOptions
	SearchFrom time.Time
	SearchTo   time.Time
	Tickers    []string
}

func (q ScanQuery) Hash() string {
	h := sha256.New()
	enc := json.NewEncoder(h)
	_ = enc.Encode(q.Segment.Ticker)
	_ = enc.Encode(q.Segment.From)
	_ = enc.Encode(q.Segment.To)
	_ = enc.Encode(q.Options)
	_ = enc.Encode(q.SearchFrom.Unix())
	_ = enc.Encode(q.SearchTo.Unix())
	_ = enc.Encode(q.Tickers)
	return hex.EncodeToString(h.Sum(nil))
}
