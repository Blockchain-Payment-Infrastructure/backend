package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			path = path + "?" + c.Request.URL.RawQuery
		}

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		logger := slog.With(
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", statusCode),
			slog.Duration("latency", latency),
			slog.String("ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
		)

		if len(c.Errors) > 0 {
			logger.Error("Request failed", slog.String("error", c.Errors.String()))
		} else {
			logger.Info("Request Handled")
		}
	}
}
