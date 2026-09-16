package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/jobman/backend/pkg/httpapi"
)

// RateLimiter is a simple in-memory per-key sliding-window limiter.
// Suitable for MVP login throttling in a single-instance deployment.
type RateLimiter struct {
	mu      sync.Mutex
	window  time.Duration
	max     int
	buckets map[string][]time.Time
	cleanup time.Duration
}

func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		window:  window,
		max:     max,
		buckets: make(map[string][]time.Time),
		cleanup: 10 * time.Minute,
	}
	go rl.periodicCleanup()
	return rl
}

func (rl *RateLimiter) Allow(key string) bool {
	now := time.Now()
	cutoff := now.Add(-rl.window)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	hits := rl.buckets[key]
	kept := hits[:0]
	for _, t := range hits {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= rl.max {
		rl.buckets[key] = kept
		return false
	}
	rl.buckets[key] = append(kept, now)
	return true
}

func (rl *RateLimiter) periodicCleanup() {
	for {
		time.Sleep(rl.cleanup)
		now := time.Now().Add(-rl.window)
		rl.mu.Lock()
		for k, hits := range rl.buckets {
			kept := hits[:0]
			for _, t := range hits {
				if t.After(now) {
					kept = append(kept, t)
				}
			}
			if len(kept) == 0 {
				delete(rl.buckets, k)
			} else {
				rl.buckets[k] = kept
			}
		}
		rl.mu.Unlock()
	}
}

func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// RateLimit throttles by key derived from the request (IP by default).
func RateLimit(limiter *RateLimiter, keyFn func(*http.Request) string) func(http.Handler) http.Handler {
	if keyFn == nil {
		keyFn = ClientIP
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow(keyFn(r)) {
				httpapi.WriteError(w, http.StatusTooManyRequests, "RATE_LIMITED", "Too many attempts, please try again later.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}