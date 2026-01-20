package scheduler

import (
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	"log/slog"
	"time"
)

type Job interface {
	Run(ctx context.Context)
}

type Scheduler struct {
	log  *slog.Logger
	job  Job
	cron *cron.Cron
}

func New(
	cronExpr string,
	timezone string,
	job Job,
) (*Scheduler, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to load location: %w", err)
	}

	c := cron.New(cron.WithLocation(loc))

	_, err = c.AddFunc(cronExpr, func() {
		job.Run(context.Background())
	})

	if err != nil {
		return nil, fmt.Errorf("failed to add job: %w", err)
	}

	return &Scheduler{
		job:  job,
		cron: c,
	}, nil
}

func (s *Scheduler) Start(ctx context.Context) {
	s.cron.Start()
	<-ctx.Done()
	s.cron.Stop()
}
