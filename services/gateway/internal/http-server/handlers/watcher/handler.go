package watcher

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/m1keee3/FinanceAnalyst/pkg/logger/sl"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/http-server/ginutils"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/http-server/handlers/watcher/mapper"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/models/watcher"
	"log/slog"
	"net/http"
)

type Client interface {
	GetStats(ctx context.Context) ([]*watcher.SegmentStats, error)
}

type Handler struct {
	log    *slog.Logger
	client Client
}

func New(log *slog.Logger, client Client) *Handler {
	return &Handler{
		log:    log,
		client: client,
	}
}

func (h *Handler) GetStats(c *gin.Context) {
	const op = "watcher.Handler.GetStats"

	log := h.log.With(slog.String("op", op))
	log.Info("get stats request")

	stats, err := h.client.GetStats(c.Request.Context())
	if err != nil {
		log.Error("failed to get stats", sl.Err(err))
		ginutils.HandleGrpcError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapper.FromStats(stats))
}
