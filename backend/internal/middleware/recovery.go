package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

type panicLog struct {
	Level     string `json:"level"`
	RequestID string `json:"request_id"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Error     string `json:"error"`
	Stack     string `json:"stack"`
}

// PanicRecovery replaces gin.Recovery() with structured JSON panic logging.
// The stack trace is logged server-side but never sent to the client.
func PanicRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()

				entry := panicLog{
					Level:     "PANIC",
					RequestID: GetRequestID(c),
					Method:    c.Request.Method,
					Path:      c.Request.URL.Path,
					Error:     fmt.Sprintf("%v", err),
					Stack:     string(stack),
				}
				if b, jsonErr := json.Marshal(entry); jsonErr == nil {
					log.Println(string(b))
				}

				// Never expose the stack trace to the client
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error_code": "INTERNAL_ERROR",
					"message":    "internal server error",
				})
			}
		}()
		c.Next()
	}
}
