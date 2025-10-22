package scanner

import (
	"github.com/m1keee3/FinanceAnalyst/common/models"
)

type StatsComputer interface {
	ComputeStats(matches []models.ChartSegment, daysToWatch int) (*models.ScanStats, error)
}
