package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

func requestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Next()
	}
}

func structuredLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"ip", c.ClientIP(),
			"request_id", c.GetString("request_id"),
		)
	}
}

// globalLimiter allows 20 requests/second with a burst of 40.
// Sufficient for an internal admin tool while preventing accidental hammering.
var globalLimiter = rate.NewLimiter(20, 40)

func rateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !globalLimiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, ErrorResponse{Error: "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

func apiKeyAuth(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-Api-Key") != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid or missing API key"})
			return
		}
		c.Next()
	}
}
