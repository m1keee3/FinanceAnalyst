package scanner

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/m1keee3/FinanceAnalyst/pkg/logger/sl"
	candleDTO "github.com/m1keee3/FinanceAnalyst/services/gateway/internal/http-server/dto/scanner/candle"
	chartDTO "github.com/m1keee3/FinanceAnalyst/services/gateway/internal/http-server/dto/scanner/chart"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/http-server/ginutils"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/http-server/handlers/scanner/mapper"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/models/scanner"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/models/scanner/candle"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/models/scanner/chart"
	"log/slog"
	"net/http"
)

type Client interface {
	FindCandleMatches(ctx context.Context, q *candle.ScanRequest) ([]*scanner.ChartSegment, error)
	FindChartMatches(ctx context.Context, q *chart.ScanRequest) ([]*scanner.ChartSegment, error)
	ComputeCandleStats(ctx context.Context, q *candle.StatsRequest) (*scanner.ScanStats, error)
	ComputeChartStats(ctx context.Context, q *chart.StatsRequest) (*scanner.ScanStats, error)
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

func (h *Handler) GetCandleMatches(c *gin.Context) {
	const op = "scanner.Handler.GetCandleMatches"

	log := h.log.With(slog.String("op", op))
	log.Info("get candle matches request")

	var dto candleDTO.ScanRequest
	if err := c.ShouldBindJSON(&dto); err != nil {
		log.Warn("invalid request", sl.Err(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	req := mapper.ToCandleScanRequest(&dto)

	matches, err := h.client.FindCandleMatches(c.Request.Context(), req)
	if err != nil {
		log.Error("failed to get candle matches", sl.Err(err))
		ginutils.HandleGrpcError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapper.FromMatches(matches))
}

func (h *Handler) GetChartMatches(c *gin.Context) {
	const op = "scanner.Handler.GetChartMatches"

	log := h.log.With(slog.String("op", op))
	log.Info("get chart matches request")

	var dto chartDTO.ScanRequest
	if err := c.ShouldBindJSON(&dto); err != nil {
		log.Warn("invalid request", sl.Err(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	req := mapper.ToChartScanRequest(&dto)

	matches, err := h.client.FindChartMatches(c.Request.Context(), req)
	if err != nil {
		log.Error("failed to get chart matches", sl.Err(err))
		ginutils.HandleGrpcError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapper.FromMatches(matches))
}

func (h *Handler) GetCandleStats(c *gin.Context) {
	const op = "scanner.Handler.GetCandleStats"

	log := h.log.With(slog.String("op", op))
	log.Info("get candle stats request")

	var dto candleDTO.StatsRequest
	if err := c.ShouldBindJSON(&dto); err != nil {
		log.Warn("invalid request", sl.Err(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	req := mapper.ToCandleStatsRequest(&dto)

	stats, err := h.client.ComputeCandleStats(c.Request.Context(), req)
	if err != nil {
		log.Error("failed to compute candle stats", sl.Err(err))
		ginutils.HandleGrpcError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapper.FromScanStats(stats))
}

func (h *Handler) GetChartStats(c *gin.Context) {
	const op = "scanner.Handler.GetChartStats"

	log := h.log.With(slog.String("op", op))
	log.Info("get chart stats request")

	var dto chartDTO.StatsRequest
	if err := c.ShouldBindJSON(&dto); err != nil {
		log.Warn("invalid request", sl.Err(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	req := mapper.ToChartStatsRequest(&dto)

	stats, err := h.client.ComputeChartStats(c.Request.Context(), req)
	if err != nil {
		log.Error("failed to compute chart stats", sl.Err(err))
		ginutils.HandleGrpcError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapper.FromScanStats(stats))
}
