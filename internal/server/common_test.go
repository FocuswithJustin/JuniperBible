package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddlewareAllowAll(t *testing.T) {
	handler := CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected CORS header to allow all origins")
	}

	if resp.Header.Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected CORS methods header")
	}

	if resp.Header.Get("Access-Control-Allow-Headers") == "" {
		t.Error("expected CORS headers header")
	}
}

func TestCORSMiddlewareWithConfigRestrictedOrigins(t *testing.T) {
	allowedOrigins := []string{"https://example.com", "https://trusted.com"}
	cfg := CORSConfig{
		AllowedOrigins: allowedOrigins,
	}

	handler := CORSMiddlewareWithConfig(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name               string
		origin             string
		expectStatus       int
		expectAllowOrigin  string
		expectCredentials  bool
	}{
		{
			name:               "allowed origin",
			origin:             "https://example.com",
			expectStatus:       http.StatusOK,
			expectAllowOrigin:  "https://example.com",
			expectCredentials:  true,
		},
		{
			name:               "another allowed origin",
			origin:             "https://trusted.com",
			expectStatus:       http.StatusOK,
			expectAllowOrigin:  "https://trusted.com",
			expectCredentials:  true,
		},
		{
			name:               "disallowed origin",
			origin:             "https://evil.com",
			expectStatus:       http.StatusOK,
			expectAllowOrigin:  "",
			expectCredentials:  false,
		},
		{
			name:               "no origin header",
			origin:             "",
			expectStatus:       http.StatusOK,
			expectAllowOrigin:  "",
			expectCredentials:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectStatus {
				t.Errorf("expected status %d, got %d", tt.expectStatus, resp.StatusCode)
			}

			allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
			if allowOrigin != tt.expectAllowOrigin {
				t.Errorf("expected Allow-Origin %q, got %q", tt.expectAllowOrigin, allowOrigin)
			}

			credentials := resp.Header.Get("Access-Control-Allow-Credentials")
			hasCredentials := credentials == "true"
			if hasCredentials != tt.expectCredentials {
				t.Errorf("expected credentials %v, got %v", tt.expectCredentials, hasCredentials)
			}
		})
	}
}

func TestCORSMiddlewareOptionsRequest(t *testing.T) {
	allowedOrigins := []string{"https://example.com"}
	cfg := CORSConfig{
		AllowedOrigins: allowedOrigins,
	}

	handler := CORSMiddlewareWithConfig(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for OPTIONS request")
	}))

	t.Run("allowed origin OPTIONS", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		req.Header.Set("Origin", "https://example.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("disallowed origin OPTIONS", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		req.Header.Set("Origin", "https://evil.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", resp.StatusCode)
		}
	})
}

func TestCORSMiddlewareEmptyConfig(t *testing.T) {
	// Empty config should behave like allow-all
	cfg := CORSConfig{}

	handler := CORSMiddlewareWithConfig(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Error("empty config should allow all origins")
	}

	// Should not set credentials with wildcard
	if resp.Header.Get("Access-Control-Allow-Credentials") != "" {
		t.Error("should not set credentials with wildcard origin")
	}
}
