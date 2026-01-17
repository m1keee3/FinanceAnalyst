package watcher

import (
	"context"
	"github.com/m1keee3/FinanceAnalyst/services/watcher/domain/models"
	"log/slog"
)

type Publisher interface {
	Publish(segStats *models.SegmentStats) error
}

type Storage interface {
	GetCurrentStats(ctx context.Context) ([]models.SegmentStats, error) // Return stats that was saved this day
	SetSegmentStats(ctx context.Context, segStats models.SegmentStats) error
}

type Service struct {
	log       *slog.Logger
	publisher Publisher
}

func (s *Service) StartScheduler(ctx context.Context) {
}

func (s *Service) runDailyScan(ctx context.Context) error {
	return nil
}
