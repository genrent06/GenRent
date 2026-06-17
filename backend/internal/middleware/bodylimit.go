package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BodyLimit rejects requests whose body exceeds maxBytes.
// Example: BodyLimit(1 << 20) limits JSON bodies to 1 MB.
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error_code": "REQUEST_TOO_LARGE",
				"message":    "request body exceeds maximum allowed size",
			})
			return
		}
		// Wrap the body reader so reading beyond the limit also fails
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
