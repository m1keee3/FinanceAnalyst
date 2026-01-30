package http_server

import (
	"github.com/gin-gonic/gin"
	"github.com/m1keee3/FinanceAnalyst/services/gateway/internal/http-server/middleware/logger"
	"log/slog"
)

type ScannerHandler interface {
	GetCandleMatches(c *gin.Context)
	GetChartMatches(c *gin.Context)
	GetCandleStats(c *gin.Context)
	GetChartStats(c *gin.Context)
}

type WatcherHandler interface {
	GetStats(c *gin.Context)
}

func NewRouter(log *slog.Logger, scanner ScannerHandler, watcher WatcherHandler) *gin.Engine {
	r := gin.New()

	r.Use(logger.New(log), gin.Recovery())

	r.GET("/scanner/candle/matches", scanner.GetCandleMatches)
	r.GET("/scanner/chart/matches", scanner.GetChartMatches)
	r.GET("/scanner/candle/stats", scanner.GetCandleStats)
	r.GET("/scanner/chart/stats", scanner.GetChartStats)

	r.GET("/watcher/stats", watcher.GetStats)

	return r
}
