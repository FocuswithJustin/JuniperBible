package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddlewareDisabled(t *testing.T) {
	// When auth is disabled, all requests should pass through
	authCfg := AuthConfig{
		Enabled: false,
		APIKey:  "",
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(authCfg, testHandler)

	req := httptest.NewRequest(http.MethodGet, "/capsules", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("expected handler to be called when auth is disabled")
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestAuthMiddlewareEnabledValidKey(t *testing.T) {
	// When auth is enabled with valid key, request should pass
	authCfg := AuthConfig{
		Enabled: true,
		APIKey:  "test-api-key-12345678",
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(authCfg, testHandler)

	req := httptest.NewRequest(http.MethodGet, "/capsules", nil)
	req.Header.Set("X-API-Key", "test-api-key-12345678")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("expected handler to be called with valid API key")
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestAuthMiddlewareEnabledMissingKey(t *testing.T) {
	// When auth is enabled without key, request should be rejected
	authCfg := AuthConfig{
		Enabled: true,
		APIKey:  "test-api-key-12345678",
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(authCfg, testHandler)

	req := httptest.NewRequest(http.MethodGet, "/capsules", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if handlerCalled {
		t.Error("expected handler not to be called without API key")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddlewareEnabledInvalidKey(t *testing.T) {
	// When auth is enabled with wrong key, request should be rejected
	authCfg := AuthConfig{
		Enabled: true,
		APIKey:  "correct-api-key-12345678",
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(authCfg, testHandler)

	req := httptest.NewRequest(http.MethodGet, "/capsules", nil)
	req.Header.Set("X-API-Key", "wrong-api-key")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if handlerCalled {
		t.Error("expected handler not to be called with invalid API key")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddlewarePublicEndpointRoot(t *testing.T) {
	// Public endpoints should always be accessible
	authCfg := AuthConfig{
		Enabled: true,
		APIKey:  "test-api-key-12345678",
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(authCfg, testHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("expected handler to be called for public endpoint /")
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestAuthMiddlewarePublicEndpointHealth(t *testing.T) {
	// Health endpoint should always be accessible
	authCfg := AuthConfig{
		Enabled: true,
		APIKey:  "test-api-key-12345678",
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(authCfg, testHandler)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("expected handler to be called for public endpoint /health")
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestIsPublicEndpoint(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/", true},
		{"/health", true},
		{"/capsules", false},
		{"/plugins", false},
		{"/formats", false},
		{"/convert", false},
		{"/capsules/test.tar.xz", false},
	}

	for _, tc := range tests {
		result := isPublicEndpoint(tc.path)
		if result != tc.expected {
			t.Errorf("isPublicEndpoint(%q) = %v, want %v", tc.path, result, tc.expected)
		}
	}
}

func TestValidateAuthConfigValid(t *testing.T) {
	tests := []struct {
		name   string
		config AuthConfig
	}{
		{
			name: "disabled auth",
			config: AuthConfig{
				Enabled: false,
				APIKey:  "",
			},
		},
		{
			name: "enabled with valid key",
			config: AuthConfig{
				Enabled: true,
				APIKey:  "valid-api-key-16chars",
			},
		},
		{
			name: "enabled with long key",
			config: AuthConfig{
				Enabled: true,
				APIKey:  "very-long-api-key-with-many-characters-for-security",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAuthConfig(tc.config)
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestValidateAuthConfigInvalid(t *testing.T) {
	tests := []struct {
		name   string
		config AuthConfig
	}{
		{
			name: "enabled without key",
			config: AuthConfig{
				Enabled: true,
				APIKey:  "",
			},
		},
		{
			name: "enabled with short key",
			config: AuthConfig{
				Enabled: true,
				APIKey:  "short",
			},
		},
		{
			name: "enabled with 15 char key",
			config: AuthConfig{
				Enabled: true,
				APIKey:  "123456789012345",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAuthConfig(tc.config)
			if err == nil {
				t.Error("expected error for invalid config")
			}
		})
	}
}

func TestGenerateAPIKeyExample(t *testing.T) {
	example := GenerateAPIKeyExample()
	if example == "" {
		t.Error("expected non-empty example string")
	}
	if len(example) < 20 {
		t.Error("expected example to contain meaningful message")
	}
}

func TestAuthMiddlewareIntegrationWithFormats(t *testing.T) {
	// Test that authenticated request can access /formats
	authCfg := AuthConfig{
		Enabled: true,
		APIKey:  "test-api-key-12345678",
	}

	middleware := AuthMiddleware(authCfg, http.HandlerFunc(handleFormats))

	req := httptest.NewRequest(http.MethodGet, "/formats", nil)
	req.Header.Set("X-API-Key", "test-api-key-12345678")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestAuthMiddlewareIntegrationWithFormatsUnauth(t *testing.T) {
	// Test that unauthenticated request is blocked from /formats
	authCfg := AuthConfig{
		Enabled: true,
		APIKey:  "test-api-key-12345678",
	}

	middleware := AuthMiddleware(authCfg, http.HandlerFunc(handleFormats))

	req := httptest.NewRequest(http.MethodGet, "/formats", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddlewareIntegrationWithHealth(t *testing.T) {
	// Test that /health is always accessible even when auth is enabled
	authCfg := AuthConfig{
		Enabled: true,
		APIKey:  "test-api-key-12345678",
	}

	middleware := AuthMiddleware(authCfg, http.HandlerFunc(handleHealth))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestAuthMiddlewareCaseSensitiveKey(t *testing.T) {
	// API keys should be case-sensitive
	authCfg := AuthConfig{
		Enabled: true,
		APIKey:  "CaseSensitiveKey123",
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(authCfg, testHandler)

	// Test with correct case
	req := httptest.NewRequest(http.MethodGet, "/capsules", nil)
	req.Header.Set("X-API-Key", "CaseSensitiveKey123")
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 with correct case, got %d", w.Code)
	}

	// Test with wrong case
	req = httptest.NewRequest(http.MethodGet, "/capsules", nil)
	req.Header.Set("X-API-Key", "casesensitivekey123")
	w = httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 with wrong case, got %d", w.Code)
	}
}

func TestAuthMiddlewareMultipleRequests(t *testing.T) {
	// Ensure middleware works correctly for multiple requests
	authCfg := AuthConfig{
		Enabled: true,
		APIKey:  "test-api-key-12345678",
	}

	callCount := 0
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(authCfg, testHandler)

	// First request with valid key
	req := httptest.NewRequest(http.MethodGet, "/capsules", nil)
	req.Header.Set("X-API-Key", "test-api-key-12345678")
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}

	// Second request without key
	req = httptest.NewRequest(http.MethodGet, "/capsules", nil)
	w = httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	if callCount != 1 {
		t.Errorf("expected still 1 call (unauthorized request shouldn't reach handler), got %d", callCount)
	}

	// Third request with valid key
	req = httptest.NewRequest(http.MethodGet, "/capsules", nil)
	req.Header.Set("X-API-Key", "test-api-key-12345678")
	w = httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}
