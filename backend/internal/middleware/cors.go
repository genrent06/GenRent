package middleware

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS configures cross-origin rules.
// allowedOrigins: comma-separated list of allowed origins, or "*" for development.
// In production, set ALLOWED_ORIGINS=https://genrent.com,https://app.genrent.com
func CORS(allowedOrigins string) gin.HandlerFunc {
	cfg := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}

	if allowedOrigins == "*" || allowedOrigins == "" {
		cfg.AllowAllOrigins = true
	} else {
		cfg.AllowOrigins = splitTrim(allowedOrigins)
	}

	return cors.New(cfg)
}

func splitTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}
