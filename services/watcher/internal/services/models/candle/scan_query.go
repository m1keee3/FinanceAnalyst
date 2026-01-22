package candle

import (
	"github.com/m1keee3/FinanceAnalyst/services/watcher/domain/models"
	"time"
)

type ScanQuery struct {
	Segment    *models.ChartSegment
	Options    *ScanOptions
	SearchFrom time.Time
	SearchTo   time.Time
	Tickers    []string
}
