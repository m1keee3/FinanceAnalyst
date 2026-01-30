package chart

import (
	"time"

	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/models/scanner"
)

type ScanRequest struct {
	Segment    *scanner.ChartSegment
	Options    *ScanOptions
	SearchFrom time.Time
	SearchTo   time.Time
	Tickers    []string
}
