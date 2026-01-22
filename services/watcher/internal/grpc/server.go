package grpc

import (
	"context"
	"github.com/m1keee3/FinanceAnalyst/services/watcher/domain/models"
	"github.com/m1keee3/FinanceAnalyst/services/watcher/internal/grpc/mapper"
	watcherv1 "github.com/m1keee3/FinanceAnalyst/services/watcher/proto-gen/watcher/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WatcherService interface {
	GetStats(ctx context.Context) ([]models.SegmentStats, error)
}

type serverAPI struct {
	watcherv1.UnimplementedWatcherServer
	watcherService WatcherService
}

func Register(grpcServer *grpc.Server, watcherService WatcherService) {
	watcherv1.RegisterWatcherServer(grpcServer, &serverAPI{watcherService: watcherService})
}

func (s *serverAPI) GetStats(ctx context.Context, request *watcherv1.GetStatsRequest) (*watcherv1.GetStatsResponse, error) {
	stats, err := s.watcherService.GetStats(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get stats")
	}

	return mapper.StatsToGetStatsResponse(stats), nil
}
