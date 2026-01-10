package api

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiterConfig holds rate limiter configuration.
type RateLimiterConfig struct {
	RequestsPerMinute int
	BurstSize         int
}

// tokenBucket implements a token bucket rate limiter.
type tokenBucket struct {
	tokens         float64
	capacity       float64
	refillRate     float64 // tokens per second
	lastRefillTime time.Time
	mu             sync.Mutex
}

// newTokenBucket creates a new token bucket with the given capacity and refill rate.
func newTokenBucket(capacity, refillRate float64) *tokenBucket {
	return &tokenBucket{
		tokens:         capacity,
		capacity:       capacity,
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
	}
}

// allow checks if a request can be allowed (returns true if token available).
func (tb *tokenBucket) allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime).Seconds()

	// Refill tokens based on time elapsed
	tb.tokens = min(tb.capacity, tb.tokens+elapsed*tb.refillRate)
	tb.lastRefillTime = now

	// Check if we have a token available
	if tb.tokens >= 1.0 {
		tb.tokens--
		return true
	}

	return false
}

// remaining returns the number of tokens remaining.
func (tb *tokenBucket) remaining() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime).Seconds()

	// Calculate current tokens
	tokens := min(tb.capacity, tb.tokens+elapsed*tb.refillRate)
	return int(tokens)
}

// reset returns the time when the bucket will be full again.
func (tb *tokenBucket) reset() time.Time {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime).Seconds()
	tokens := min(tb.capacity, tb.tokens+elapsed*tb.refillRate)

	if tokens >= tb.capacity {
		return now
	}

	tokensNeeded := tb.capacity - tokens
	secondsUntilFull := tokensNeeded / tb.refillRate
	return now.Add(time.Duration(secondsUntilFull * float64(time.Second)))
}

// RateLimiter manages per-IP rate limiting.
type RateLimiter struct {
	buckets    map[string]*tokenBucket
	config     RateLimiterConfig
	mu         sync.RWMutex
	cleanupTTL time.Duration
}

// NewRateLimiter creates a new rate limiter with the given configuration.
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	rl := &RateLimiter{
		buckets:    make(map[string]*tokenBucket),
		config:     config,
		cleanupTTL: 5 * time.Minute,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// getBucket returns the token bucket for a given IP, creating if necessary.
func (rl *RateLimiter) getBucket(ip string) *tokenBucket {
	rl.mu.RLock()
	bucket, exists := rl.buckets[ip]
	rl.mu.RUnlock()

	if exists {
		return bucket
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	if bucket, exists := rl.buckets[ip]; exists {
		return bucket
	}

	// Create new bucket
	// Convert requests per minute to tokens per second
	refillRate := float64(rl.config.RequestsPerMinute) / 60.0
	capacity := float64(rl.config.BurstSize)

	bucket = newTokenBucket(capacity, refillRate)
	rl.buckets[ip] = bucket

	return bucket
}

// cleanup periodically removes stale buckets.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, bucket := range rl.buckets {
			// Remove buckets that have been idle for longer than TTL
			if now.Sub(bucket.lastRefillTime) > rl.cleanupTTL {
				delete(rl.buckets, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if a request from the given IP should be allowed.
func (rl *RateLimiter) Allow(ip string) bool {
	bucket := rl.getBucket(ip)
	return bucket.allow()
}

// Remaining returns the number of remaining requests for the given IP.
func (rl *RateLimiter) Remaining(ip string) int {
	bucket := rl.getBucket(ip)
	return bucket.remaining()
}

// Reset returns the reset time for the given IP.
func (rl *RateLimiter) Reset(ip string) time.Time {
	bucket := rl.getBucket(ip)
	return bucket.reset()
}

// Middleware returns an HTTP middleware that applies rate limiting.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)

		bucket := rl.getBucket(ip)
		remaining := bucket.remaining()
		reset := bucket.reset()

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.config.RequestsPerMinute))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", reset.Unix()))

		if !bucket.allow() {
			retryAfter := int(time.Until(reset).Seconds()) + 1
			w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))

			respondError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED",
				fmt.Sprintf("Rate limit exceeded. Try again in %d seconds.", retryAfter))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts the client IP address from the request.
// Checks X-Forwarded-For and X-Real-IP headers before falling back to RemoteAddr.
// SEC-001 FIX: Properly parses X-Forwarded-For header and validates IP format.
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	// Format: X-Forwarded-For: client, proxy1, proxy2
	// Take the leftmost (client) IP and validate it
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Split by comma and take the first (leftmost) IP
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			clientIP := strings.TrimSpace(ips[0])
			// Validate IP format before using it
			if isValidIP(clientIP) {
				return clientIP
			}
		}
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		// Validate IP format
		realIP = strings.TrimSpace(realIP)
		if isValidIP(realIP) {
			return realIP
		}
	}

	// Fall back to RemoteAddr
	// RemoteAddr is in format "IP:port", strip the port
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If SplitHostPort fails, try to use RemoteAddr directly
		// This handles cases where RemoteAddr is just an IP without port
		ip = r.RemoteAddr
	}

	// Final validation
	if isValidIP(ip) {
		return ip
	}

	// If all else fails, return a safe default
	return "unknown"
}

// isValidIP checks if a string is a valid IPv4 or IPv6 address.
func isValidIP(ipStr string) bool {
	return net.ParseIP(ipStr) != nil
}

// min returns the minimum of two float64 values.
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
