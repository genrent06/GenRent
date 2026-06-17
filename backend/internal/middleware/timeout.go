package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestTimeout aborts any request that takes longer than the given duration.
// Long-running requests (stuck DB queries, slow uploads) are cut off cleanly.
func RequestTimeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		done := make(chan struct{})
		go func() {
			c.Next()
			close(done)
		}()
		select {
		case <-done:
			// completed normally
		case <-time.After(d):
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
				"error": "request timed out — please try again",
			})
		}
	}
}
