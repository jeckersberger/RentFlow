package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeckersberger/rentflow/pkg/common/config"
)

func TestCORSMiddleware_AllowsConfiguredOrigin(t *testing.T) {
	cfg := &config.Config{
		CORSAllowedOrigins: []string{"https://example.com"},
		CORSAllowedMethods: []string{"GET", "POST"},
		CORSAllowedHeaders: []string{"Content-Type", "Authorization"},
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CORSMiddleware(cfg)
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("expected Access-Control-Allow-Origin 'https://example.com', got '%s'", w.Header().Get("Access-Control-Allow-Origin"))
	}

	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("expected Access-Control-Allow-Credentials 'true', got '%s'", w.Header().Get("Access-Control-Allow-Credentials"))
	}
}

func TestCORSMiddleware_DeniesUnallowedOrigin(t *testing.T) {
	cfg := &config.Config{
		CORSAllowedOrigins: []string{"https://example.com"},
		CORSAllowedMethods: []string{"GET", "POST"},
		CORSAllowedHeaders: []string{"Content-Type"},
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CORSMiddleware(cfg)
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://unauthorized.com")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("expected no Access-Control-Allow-Origin header for unauthorized origin, got '%s'", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSMiddleware_AllowWildcard(t *testing.T) {
	cfg := &config.Config{
		CORSAllowedOrigins: []string{"*"},
		CORSAllowedMethods: []string{"GET", "POST"},
		CORSAllowedHeaders: []string{"Content-Type"},
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CORSMiddleware(cfg)
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://any-origin.com")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "https://any-origin.com" {
		t.Errorf("expected Access-Control-Allow-Origin with wildcard, got '%s'", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSMiddleware_PreflightRequest(t *testing.T) {
	cfg := &config.Config{
		CORSAllowedOrigins: []string{"https://example.com"},
		CORSAllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		CORSAllowedHeaders: []string{"Content-Type", "Authorization"},
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CORSMiddleware(cfg)
	handler := middleware(testHandler)

	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "https://example.com")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for preflight, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("expected Access-Control-Allow-Origin header in preflight response")
	}
}

func TestCORSMiddleware_HeadersSet(t *testing.T) {
	cfg := &config.Config{
		CORSAllowedOrigins: []string{"https://example.com"},
		CORSAllowedMethods: []string{"GET", "POST", "DELETE"},
		CORSAllowedHeaders: []string{"Content-Type", "X-Custom-Header"},
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CORSMiddleware(cfg)
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Methods") != "GET, POST, DELETE" {
		t.Errorf("expected Access-Control-Allow-Methods header set correctly, got '%s'", w.Header().Get("Access-Control-Allow-Methods"))
	}

	if w.Header().Get("Access-Control-Allow-Headers") != "Content-Type, X-Custom-Header" {
		t.Errorf("expected Access-Control-Allow-Headers header set correctly, got '%s'", w.Header().Get("Access-Control-Allow-Headers"))
	}
}

func TestCORSMiddleware_ExposeHeaders(t *testing.T) {
	cfg := &config.Config{
		CORSAllowedOrigins: []string{"*"},
		CORSAllowedMethods: []string{"GET"},
		CORSAllowedHeaders: []string{"Content-Type"},
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CORSMiddleware(cfg)
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Expose-Headers") != "Content-Length, X-Request-ID" {
		t.Errorf("expected Access-Control-Expose-Headers header, got '%s'", w.Header().Get("Access-Control-Expose-Headers"))
	}
}

func TestSimpleCORSMiddleware_DefaultIsRestrictive(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "")

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := SimpleCORSMiddleware()
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://any-origin.com")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no Access-Control-Allow-Origin by default, got '%s'", got)
	}
}

func TestSimpleCORSMiddleware_Methods(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := SimpleCORSMiddleware()
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	methods := w.Header().Get("Access-Control-Allow-Methods")
	if methods == "" {
		t.Error("expected Access-Control-Allow-Methods header to be set")
	}

	expectedMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	for _, method := range expectedMethods {
		if !containsMethod(methods, method) {
			t.Errorf("expected method '%s' in '%s'", method, methods)
		}
	}
}

func TestSimpleCORSMiddleware_Headers(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := SimpleCORSMiddleware()
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	headers := w.Header().Get("Access-Control-Allow-Headers")
	if headers == "" {
		t.Error("expected Access-Control-Allow-Headers header to be set")
	}

	expectedHeaders := []string{"Content-Type", "Authorization"}
	for _, header := range expectedHeaders {
		if !containsHeader(headers, header) {
			t.Errorf("expected header '%s' in '%s'", header, headers)
		}
	}
}

func TestSimpleCORSMiddleware_PreflightRequest(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := SimpleCORSMiddleware()
	handler := middleware(testHandler)

	req := httptest.NewRequest("OPTIONS", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for preflight, got %d", w.Code)
	}
}

func TestRestrictiveCORSMiddleware_AllowedOrigin(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RestrictiveCORSMiddleware("https://app.example.com", "https://admin.example.com")
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://app.example.com")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "https://app.example.com" {
		t.Errorf("expected Access-Control-Allow-Origin 'https://app.example.com', got '%s'", w.Header().Get("Access-Control-Allow-Origin"))
	}

	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("expected Access-Control-Allow-Credentials 'true'")
	}
}

func TestRestrictiveCORSMiddleware_DeniedOrigin(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RestrictiveCORSMiddleware("https://app.example.com")
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://unauthorized.com")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("expected no Access-Control-Allow-Origin header for unauthorized origin")
	}
}

func TestIsOriginAllowed_Wildcard(t *testing.T) {
	allowed := isOriginAllowed("https://any-origin.com", []string{"*"})

	if !allowed {
		t.Error("expected wildcard '*' to allow any origin")
	}
}

func TestIsOriginAllowed_ExactMatch(t *testing.T) {
	allowed := isOriginAllowed("https://example.com", []string{"https://example.com"})

	if !allowed {
		t.Error("expected exact match to be allowed")
	}
}

func TestIsOriginAllowed_WildcardSubdomain(t *testing.T) {
	allowed := isOriginAllowed("https://app.example.com", []string{"*.example.com"})

	if !allowed {
		t.Error("expected wildcard subdomain '*.example.com' to allow 'app.example.com'")
	}
}

func TestIsOriginAllowed_WildcardSubdomain_NoMatch(t *testing.T) {
	allowed := isOriginAllowed("https://other.net", []string{"*.example.com"})

	if allowed {
		t.Error("expected wildcard subdomain not to match different domain")
	}
}

func TestIsOriginAllowed_EmptyOrigin(t *testing.T) {
	allowed := isOriginAllowed("", []string{"*", "https://example.com"})

	if allowed {
		t.Error("expected empty origin to be rejected")
	}
}

func TestIsOriginAllowed_MultipleOrigins(t *testing.T) {
	tests := []struct {
		name     string
		origin   string
		allowed  []string
		expected bool
	}{
		{"First match", "https://first.com", []string{"https://first.com", "https://second.com"}, true},
		{"Second match", "https://second.com", []string{"https://first.com", "https://second.com"}, true},
		{"No match", "https://third.com", []string{"https://first.com", "https://second.com"}, false},
	}

	for _, tt := range tests {
		result := isOriginAllowed(tt.origin, tt.allowed)
		if result != tt.expected {
			t.Errorf("isOriginAllowed(%s): expected %v, got %v", tt.name, tt.expected, result)
		}
	}
}

// Helper functions for testing

func containsMethod(methods, method string) bool {
	parts := splitAndTrim(methods)
	for _, m := range parts {
		if m == method {
			return true
		}
	}
	return false
}

func containsHeader(headers, header string) bool {
	parts := splitAndTrim(headers)
	for _, h := range parts {
		if h == header {
			return true
		}
	}
	return false
}

func splitAndTrim(s string) []string {
	var parts []string
	for _, p := range split(s, ',') {
		if trimmed := trim(p); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func split(s string, sep byte) []string {
	var parts []string
	var current string
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(s[i])
		}
	}
	parts = append(parts, current)
	return parts
}

func trim(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}

	return s[start:end]
}
