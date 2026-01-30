package logger

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"time"
)

func New(log *slog.Logger) gin.HandlerFunc {
	log = log.With(
		slog.String("component", "middleware/logger"),
	)

	return func(c *gin.Context) {
		start := time.Now()

		entry := log.With(
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("client_ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
		)

		if rid := c.GetString("request_id"); rid != "" {
			entry = entry.With(slog.String("request_id", rid))
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		size := c.Writer.Size() // bytes written

		entry.Info("request completed",
			slog.Int("status", status),
			slog.Int("bytes", size),
			slog.String("duration", latency.String()),
		)
	}
}
