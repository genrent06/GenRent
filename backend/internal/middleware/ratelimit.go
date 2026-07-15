package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipTracker struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	maxReqs  int
	window   time.Duration
}

func newIPTracker(maxReqs int, window time.Duration) *ipTracker {
	t := &ipTracker{
		requests: make(map[string][]time.Time),
		maxReqs:  maxReqs,
		window:   window,
	}
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			t.cleanup()
		}
	}()
	return t
}

func (t *ipTracker) cleanup() {
	t.mu.Lock()
	defer t.mu.Unlock()
	cutoff := time.Now().Add(-t.window)
	for ip, times := range t.requests {
		var valid []time.Time
		for _, ts := range times {
			if ts.After(cutoff) {
				valid = append(valid, ts)
			}
		}
		if len(valid) == 0 {
			delete(t.requests, ip)
		} else {
			t.requests[ip] = valid
		}
	}
}

func (t *ipTracker) allow(ip string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-t.window)
	var valid []time.Time
	for _, ts := range t.requests[ip] {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	if len(valid) >= t.maxReqs {
		t.requests[ip] = valid
		return false
	}
	t.requests[ip] = append(valid, now)
	return true
}

// RateLimit returns a middleware that allows maxReqs per windowSec seconds per IP
func RateLimit(maxReqs int, windowSec int) gin.HandlerFunc {
	tracker := newIPTracker(maxReqs, time.Duration(windowSec)*time.Second)
	return func(c *gin.Context) {
		if !tracker.allow(c.ClientIP()) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":               "too many requests — please slow down",
				"retry_after_seconds": windowSec,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
