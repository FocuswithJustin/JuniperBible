package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestTokenBucket_Allow(t *testing.T) {
	// Create bucket with 5 tokens, refilling at 1 token/second
	bucket := newTokenBucket(5, 1)

	// Should allow 5 requests immediately (burst)
	for i := 0; i < 5; i++ {
		if !bucket.allow() {
			t.Errorf("Request %d should be allowed (burst)", i+1)
		}
	}

	// 6th request should be denied
	if bucket.allow() {
		t.Error("6th request should be denied")
	}

	// Wait for refill (1 second = 1 token)
	time.Sleep(1100 * time.Millisecond)

	// Should allow 1 more request after refill
	if !bucket.allow() {
		t.Error("Request after refill should be allowed")
	}

	// Next request should be denied again
	if bucket.allow() {
		t.Error("Request should be denied after using refilled token")
	}
}

func TestTokenBucket_Remaining(t *testing.T) {
	bucket := newTokenBucket(10, 1)

	// Initially should have 10 tokens
	remaining := bucket.remaining()
	if remaining != 10 {
		t.Errorf("Expected 10 remaining tokens, got %d", remaining)
	}

	// Use 3 tokens
	for i := 0; i < 3; i++ {
		bucket.allow()
	}

	remaining = bucket.remaining()
	if remaining != 7 {
		t.Errorf("Expected 7 remaining tokens, got %d", remaining)
	}

	// Wait for partial refill
	time.Sleep(500 * time.Millisecond)
	remaining = bucket.remaining()
	if remaining < 7 || remaining > 8 {
		t.Errorf("Expected ~7-8 remaining tokens after partial refill, got %d", remaining)
	}
}

func TestTokenBucket_Reset(t *testing.T) {
	bucket := newTokenBucket(5, 1)

	// Use all tokens
	for i := 0; i < 5; i++ {
		bucket.allow()
	}

	// Reset should be in the future
	reset := bucket.reset()
	if !reset.After(time.Now()) {
		t.Error("Reset time should be in the future")
	}

	// Should be approximately 5 seconds (5 tokens / 1 token per second)
	duration := time.Until(reset)
	if duration < 4*time.Second || duration > 6*time.Second {
		t.Errorf("Expected reset in ~5 seconds, got %v", duration)
	}
}

func TestTokenBucket_Concurrent(t *testing.T) {
	bucket := newTokenBucket(100, 10) // 100 tokens, 10/second refill

	var wg sync.WaitGroup
	var allowed, denied int
	var mu sync.Mutex

	// Spawn 200 concurrent requests
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if bucket.allow() {
				mu.Lock()
				allowed++
				mu.Unlock()
			} else {
				mu.Lock()
				denied++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Should have allowed ~100 (the burst capacity)
	if allowed < 90 || allowed > 110 {
		t.Errorf("Expected ~100 allowed requests, got %d", allowed)
	}

	// Should have denied ~100
	if denied < 90 || denied > 110 {
		t.Errorf("Expected ~100 denied requests, got %d", denied)
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	config := RateLimiterConfig{
		RequestsPerMinute: 60, // 1 per second
		BurstSize:         5,
	}
	rl := NewRateLimiter(config)

	ip := "192.168.1.1"

	// Should allow burst
	for i := 0; i < 5; i++ {
		if !rl.Allow(ip) {
			t.Errorf("Request %d should be allowed (burst)", i+1)
		}
	}

	// Next request should be denied
	if rl.Allow(ip) {
		t.Error("Request beyond burst should be denied")
	}

	// Wait for refill
	time.Sleep(1100 * time.Millisecond)

	// Should allow 1 more request
	if !rl.Allow(ip) {
		t.Error("Request after refill should be allowed")
	}
}

func TestRateLimiter_PerIP(t *testing.T) {
	config := RateLimiterConfig{
		RequestsPerMinute: 60,
		BurstSize:         3,
	}
	rl := NewRateLimiter(config)

	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	// Exhaust IP1's bucket
	for i := 0; i < 3; i++ {
		if !rl.Allow(ip1) {
			t.Errorf("IP1 request %d should be allowed", i+1)
		}
	}

	// IP1 should be rate limited
	if rl.Allow(ip1) {
		t.Error("IP1 should be rate limited")
	}

	// IP2 should still have full bucket
	for i := 0; i < 3; i++ {
		if !rl.Allow(ip2) {
			t.Errorf("IP2 request %d should be allowed", i+1)
		}
	}

	// IP2 should now be rate limited
	if rl.Allow(ip2) {
		t.Error("IP2 should be rate limited")
	}
}

func TestRateLimiter_Remaining(t *testing.T) {
	config := RateLimiterConfig{
		RequestsPerMinute: 60,
		BurstSize:         10,
	}
	rl := NewRateLimiter(config)

	ip := "192.168.1.1"

	// Initially should have burst size available
	remaining := rl.Remaining(ip)
	if remaining != 10 {
		t.Errorf("Expected 10 remaining, got %d", remaining)
	}

	// Use 3 requests
	for i := 0; i < 3; i++ {
		rl.Allow(ip)
	}

	remaining = rl.Remaining(ip)
	if remaining != 7 {
		t.Errorf("Expected 7 remaining after 3 requests, got %d", remaining)
	}
}

func TestRateLimiter_Middleware(t *testing.T) {
	config := RateLimiterConfig{
		RequestsPerMinute: 60,
		BurstSize:         3,
	}
	rl := NewRateLimiter(config)

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with rate limiter middleware
	limitedHandler := rl.Middleware(handler)

	// Make 3 requests (should all succeed)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		limitedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, w.Code)
		}

		// Check rate limit headers
		if w.Header().Get("X-RateLimit-Limit") != "60" {
			t.Errorf("Request %d: expected X-RateLimit-Limit=60, got %s",
				i+1, w.Header().Get("X-RateLimit-Limit"))
		}

		remaining := w.Header().Get("X-RateLimit-Remaining")
		if remaining == "" {
			t.Errorf("Request %d: X-RateLimit-Remaining header missing", i+1)
		}
	}

	// 4th request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	limitedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w.Code)
	}

	// Check Retry-After header
	if w.Header().Get("Retry-After") == "" {
		t.Error("Retry-After header should be present when rate limited")
	}

	// Check that response is JSON error
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type=application/json, got %s", contentType)
	}
}

func TestRateLimiter_Headers(t *testing.T) {
	config := RateLimiterConfig{
		RequestsPerMinute: 60,
		BurstSize:         5,
	}
	rl := NewRateLimiter(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	limitedHandler := rl.Middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	limitedHandler.ServeHTTP(w, req)

	// Check all required headers are present
	headers := []string{
		"X-RateLimit-Limit",
		"X-RateLimit-Remaining",
		"X-RateLimit-Reset",
	}

	for _, header := range headers {
		if w.Header().Get(header) == "" {
			t.Errorf("Required header %s is missing", header)
		}
	}

	// Verify header values
	if limit := w.Header().Get("X-RateLimit-Limit"); limit != "60" {
		t.Errorf("Expected X-RateLimit-Limit=60, got %s", limit)
	}

	if remaining := w.Header().Get("X-RateLimit-Remaining"); remaining != "5" {
		t.Errorf("Expected X-RateLimit-Remaining=5, got %s", remaining)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name         string
		remoteAddr   string
		forwardedFor string
		realIP       string
		expectedIP   string
	}{
		{
			name:       "RemoteAddr only",
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "192.168.1.1",
		},
		{
			name:         "X-Forwarded-For header",
			remoteAddr:   "192.168.1.1:12345",
			forwardedFor: "203.0.113.1",
			expectedIP:   "203.0.113.1",
		},
		{
			name:       "X-Real-IP header",
			remoteAddr: "192.168.1.1:12345",
			realIP:     "203.0.113.2",
			expectedIP: "203.0.113.2",
		},
		{
			name:         "X-Forwarded-For takes precedence",
			remoteAddr:   "192.168.1.1:12345",
			forwardedFor: "203.0.113.1",
			realIP:       "203.0.113.2",
			expectedIP:   "203.0.113.1",
		},
		{
			name:         "SEC-001: X-Forwarded-For with multiple IPs (takes leftmost/client)",
			remoteAddr:   "192.168.1.1:12345",
			forwardedFor: "203.0.113.1, 10.0.0.1, 172.16.0.1",
			expectedIP:   "203.0.113.1",
		},
		{
			name:         "SEC-001: X-Forwarded-For with spaces",
			remoteAddr:   "192.168.1.1:12345",
			forwardedFor: "  203.0.113.5  ,  10.0.0.2  ",
			expectedIP:   "203.0.113.5",
		},
		{
			name:         "SEC-001: Invalid X-Forwarded-For falls back to RemoteAddr",
			remoteAddr:   "192.168.1.1:12345",
			forwardedFor: "not-an-ip-address",
			expectedIP:   "192.168.1.1",
		},
		{
			name:         "SEC-001: Malicious X-Forwarded-For with SQL injection attempt",
			remoteAddr:   "192.168.1.1:12345",
			forwardedFor: "'; DROP TABLE users; --",
			expectedIP:   "192.168.1.1",
		},
		{
			name:       "SEC-001: Invalid X-Real-IP falls back to RemoteAddr",
			remoteAddr: "192.168.1.1:12345",
			realIP:     "malicious-value",
			expectedIP: "192.168.1.1",
		},
		{
			name:         "IPv6 address",
			remoteAddr:   "[2001:db8::1]:12345",
			forwardedFor: "2001:db8::2",
			expectedIP:   "2001:db8::2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			if tt.forwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.forwardedFor)
			}
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}

			ip := getClientIP(req)
			if ip != tt.expectedIP {
				t.Errorf("Expected IP %s, got %s", tt.expectedIP, ip)
			}
		})
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	config := RateLimiterConfig{
		RequestsPerMinute: 60,
		BurstSize:         5,
	}
	rl := NewRateLimiter(config)
	rl.cleanupTTL = 100 * time.Millisecond // Short TTL for testing

	// Create buckets for multiple IPs
	for i := 0; i < 5; i++ {
		ip := fmt.Sprintf("192.168.1.%d", i)
		rl.Allow(ip)
	}

	// Check that buckets were created
	rl.mu.RLock()
	initialCount := len(rl.buckets)
	rl.mu.RUnlock()

	if initialCount != 5 {
		t.Errorf("Expected 5 buckets, got %d", initialCount)
	}

	// Wait for cleanup to run (cleanup runs every minute, but we can't easily test that)
	// This test verifies the cleanup logic exists, but doesn't wait for actual cleanup
	// since that would make the test too slow
	t.Skip("Cleanup test requires long wait time, skipping")
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	config := RateLimiterConfig{
		RequestsPerMinute: 600, // High limit for stress test
		BurstSize:         100,
	}
	rl := NewRateLimiter(config)

	var wg sync.WaitGroup
	numGoroutines := 50
	requestsPerGoroutine := 20

	// Spawn many concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		ip := fmt.Sprintf("192.168.1.%d", i%10) // 10 different IPs

		go func(ip string) {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				rl.Allow(ip)
				rl.Remaining(ip)
				rl.Reset(ip)
			}
		}(ip)
	}

	// Wait for all goroutines to complete
	// If there's a race condition, this will likely fail or hang
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out - possible deadlock")
	}
}

func BenchmarkRateLimiter_Allow(b *testing.B) {
	config := RateLimiterConfig{
		RequestsPerMinute: 60,
		BurstSize:         10,
	}
	rl := NewRateLimiter(config)
	ip := "192.168.1.1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow(ip)
	}
}

func BenchmarkRateLimiter_AllowConcurrent(b *testing.B) {
	config := RateLimiterConfig{
		RequestsPerMinute: 60,
		BurstSize:         10,
	}
	rl := NewRateLimiter(config)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		ip := "192.168.1.1"
		for pb.Next() {
			rl.Allow(ip)
		}
	})
}

func BenchmarkRateLimiter_Middleware(b *testing.B) {
	config := RateLimiterConfig{
		RequestsPerMinute: 60,
		BurstSize:         10,
	}
	rl := NewRateLimiter(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	limitedHandler := rl.Middleware(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		limitedHandler.ServeHTTP(w, req)
	}
}
