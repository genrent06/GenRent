package middleware

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

type requestLog struct {
	Level      string `json:"level"`
	RequestID  string `json:"request_id"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Status     int    `json:"status"`
	DurationMS int64  `json:"duration_ms"`
	IP         string `json:"ip"`
}

// RequestMonitor logs each request as a structured JSON line.
// Levels: INFO (normal), SLOW (>500ms), ERROR (5xx).
// Example: {"level":"INFO","request_id":"a1b2c3d4","method":"POST","path":"/api/v1/bookings","status":201,"duration_ms":18,"ip":"127.0.0.1"}
func RequestMonitor() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		level := "INFO"
		if duration > 500*time.Millisecond {
			level = "SLOW"
		}
		if status >= 500 {
			level = "ERROR"
		}

		entry := requestLog{
			Level:      level,
			RequestID:  GetRequestID(c),
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			Status:     status,
			DurationMS: duration.Milliseconds(),
			IP:         c.ClientIP(),
		}

		if b, err := json.Marshal(entry); err == nil {
			log.Println(string(b))
		}
	}
}
