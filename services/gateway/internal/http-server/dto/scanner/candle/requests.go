package candle

import (
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/http-server/dto/scanner"
	"time"
)

type StatsRequest struct {
	ScanRequest *ScanRequest `json:"scan_request"`
	DaysToWatch int          `json:"days_to_watch"`
}

type ScanRequest struct {
	Segment    *scanner.ChartSegment `json:"segment"`
	Options    *ScanOptions          `json:"scan_options"`
	SearchFrom time.Time             `json:"search_from"`
	SearchTo   time.Time             `json:"search_to"`
	Tickers    []string              `json:"tickers"`
}

type ScanOptions struct {
	TailLen         int     `json:"tail_len"`
	BodyTolerance   float64 `json:"body_tolerance"`
	ShadowTolerance float64 `json:"shadow_tolerance"`
}
