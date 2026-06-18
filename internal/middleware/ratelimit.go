package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type loginLimiter struct {
	mu      sync.Mutex
	attempt map[string][]time.Time
	limit   int
	window  time.Duration
}

func NewLoginRateLimiter(limit int, window time.Duration) *loginLimiter {
	return &loginLimiter{
		attempt: make(map[string][]time.Time),
		limit:   limit,
		window:  window,
	}
}

func (l *loginLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := clientIP(r)
		if !l.allow(key) {
			writeJSON(w, http.StatusTooManyRequests, map[string]string{
				"error": "Too many login attempts. Try again later.",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (l *loginLimiter) allow(key string) bool {
	now := time.Now()
	cutoff := now.Add(-l.window)

	l.mu.Lock()
	defer l.mu.Unlock()

	times := l.attempt[key]
	filtered := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) >= l.limit {
		l.attempt[key] = filtered
		return false
	}

	l.attempt[key] = append(filtered, now)
	return true
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
