package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/varun-2122/flashcart/internal/logger"
	"github.com/varun-2122/flashcart/internal/metrics"
	"github.com/varun-2122/flashcart/internal/response"
)

// RateLimit returns a sliding-window rate limiter middleware backed by Redis.
//
// Algorithm: Redis INCR + EXPIRE per (IP + route) key.
//   - rdb:         underlying Redis client (nil = fail open).
//   - maxRequests: maximum requests allowed in the window.
//   - window:      sliding window duration (e.g. time.Minute).
//   - route:       label for Prometheus + Redis key namespacing.
//
// Adds X-RateLimit-Limit, X-RateLimit-Remaining, Retry-After response headers.
// If Redis is unavailable the middleware fails open (allows requests through).
func RateLimit(rdb *redis.Client, maxRequests int, window time.Duration, route string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if rdb == nil {
				// Redis unavailable — fail open so service keeps running
				next.ServeHTTP(w, r)
				return
			}

			ip := realIP(r)
			key := fmt.Sprintf("rl:%s:%s", route, ip)
			ctx := r.Context()

			// Increment counter atomically
			count, err := rdb.Incr(ctx, key).Result()
			if err != nil {
				logger.Warn(ctx, "rate limiter redis error, failing open",
					"route", route,
					"ip", ip,
					"error", err.Error(),
				)
				next.ServeHTTP(w, r)
				return
			}

			// Set expiry only on first request in window (avoids resetting on every hit)
			if count == 1 {
				rdb.Expire(ctx, key, window)
			}

			remaining := maxRequests - int(count)
			if remaining < 0 {
				remaining = 0
			}

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(maxRequests))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

			if count > int64(maxRequests) {
				retryAfter := int(window.Seconds())
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))

				metrics.RateLimitHits.WithLabelValues(route).Inc()
				logger.Warn(ctx, "rate limit exceeded",
					"route", route,
					"ip", ip,
					"count", count,
					"limit", maxRequests,
				)

				response.TooManyRequests(w, "Rate limit exceeded. Please wait before retrying.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// realIP extracts the real client IP respecting common proxy headers.
func realIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	// Strip port from RemoteAddr
	addr := r.RemoteAddr
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}
	return addr
}
