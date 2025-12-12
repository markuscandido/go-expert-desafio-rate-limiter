package main_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/markuscandido/go-expert-desafio-rate-limiter/internal/config"
	"github.com/markuscandido/go-expert-desafio-rate-limiter/internal/limiter"
	"github.com/markuscandido/go-expert-desafio-rate-limiter/internal/middleware"
	"github.com/markuscandido/go-expert-desafio-rate-limiter/internal/storage"
)

// MockStorageForIntegration for integration testing
type MockStorageForIntegration struct {
	data    map[string]*storage.LimiterData
	blocked map[string]bool
}

func NewMockStorageForIntegration() *MockStorageForIntegration {
	return &MockStorageForIntegration{
		data:    make(map[string]*storage.LimiterData),
		blocked: make(map[string]bool),
	}
}

func (m *MockStorageForIntegration) CheckAndIncrement(ctx context.Context, key string, maxRequests int, windowSeconds int) (bool, error) {
	if m.blocked[key] {
		return false, nil
	}

	if m.data[key] == nil {
		m.data[key] = &storage.LimiterData{
			Count:     0,
			ExpiresAt: time.Now().Add(time.Duration(windowSeconds) * time.Second),
		}
	}

	if time.Now().After(m.data[key].ExpiresAt) {
		m.data[key].Count = 0
		m.data[key].ExpiresAt = time.Now().Add(time.Duration(windowSeconds) * time.Second)
	}

	m.data[key].Count++
	return m.data[key].Count <= maxRequests, nil
}

func (m *MockStorageForIntegration) IsBlocked(ctx context.Context, key string) (bool, error) {
	return m.blocked[key], nil
}

func (m *MockStorageForIntegration) Block(ctx context.Context, key string, durationSeconds int) error {
	m.blocked[key] = true
	return nil
}

func (m *MockStorageForIntegration) Reset(ctx context.Context, key string) error {
	delete(m.data, key)
	delete(m.blocked, key)
	return nil
}

func (m *MockStorageForIntegration) GetData(ctx context.Context, key string) (*storage.LimiterData, error) {
	return m.data[key], nil
}

func (m *MockStorageForIntegration) Close() error {
	return nil
}

// Integration Tests

func TestIPRateLimitingIntegration(t *testing.T) {
	storage := NewMockStorageForIntegration()
	cfg := &config.RateLimiterConfig{
		MaxRequestsIP:    3,
		BlockDurationIP:  60,
		EnableIPLimit:    true,
		EnableTokenLimit: false,
	}

	limiter := limiter.NewRateLimiter(storage, cfg)
	middleware := middleware.NewRateLimiterMiddleware(limiter)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	wrappedHandler := middleware.Handler(handler)

	// Make 3 allowed requests
	for i := 1; i <= 3; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.100:1234"
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d failed: expected 200, got %d", i, w.Code)
		}
	}

	// 4th request should be blocked
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.100:1234"
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("4th request failed: expected 429, got %d", w.Code)
	}
}

func TestTokenRateLimitingIntegration(t *testing.T) {
	storage := NewMockStorageForIntegration()
	cfg := &config.RateLimiterConfig{
		MaxRequestsToken:   2,
		BlockDurationToken: 60,
		EnableIPLimit:      false,
		EnableTokenLimit:   true,
	}

	limiter := limiter.NewRateLimiter(storage, cfg)
	middleware := middleware.NewRateLimiterMiddleware(limiter)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	wrappedHandler := middleware.Handler(handler)

	// Make 2 allowed requests with token
	for i := 1; i <= 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("API_KEY", "premium-token")
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d failed: expected 200, got %d", i, w.Code)
		}
	}

	// 3rd request should be blocked
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("API_KEY", "premium-token")
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("3rd request failed: expected 429, got %d", w.Code)
	}
}

func TestTokenPrecedenceIntegration(t *testing.T) {
	storage := NewMockStorageForIntegration()
	cfg := &config.RateLimiterConfig{
		MaxRequestsIP:      2,
		BlockDurationIP:    60,
		EnableIPLimit:      true,
		MaxRequestsToken:   5,
		BlockDurationToken: 60,
		EnableTokenLimit:   true,
	}

	limiter := limiter.NewRateLimiter(storage, cfg)
	middleware := middleware.NewRateLimiterMiddleware(limiter)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	wrappedHandler := middleware.Handler(handler)

	ip := "192.168.1.50:1234"
	token := "premium-token"

	// Exhaust IP limit (2 requests)
	for i := 1; i <= 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = ip
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d failed: expected 200, got %d", i, w.Code)
		}
	}

	// Request without token should be blocked
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = ip
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("3rd request without token failed: expected 429, got %d", w.Code)
	}

	// But request with token should be allowed (token takes precedence)
	storage.Reset(context.Background(), "ip:"+ip)
	req = httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = ip
	req.Header.Set("API_KEY", token)
	w = httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Request with token failed: expected 200, got %d", w.Code)
	}
}

func TestDifferentIPsAreIndependent(t *testing.T) {
	storage := NewMockStorageForIntegration()
	cfg := &config.RateLimiterConfig{
		MaxRequestsIP:    2,
		BlockDurationIP:  60,
		EnableIPLimit:    true,
		EnableTokenLimit: false,
	}

	limiter := limiter.NewRateLimiter(storage, cfg)
	middleware := middleware.NewRateLimiterMiddleware(limiter)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	wrappedHandler := middleware.Handler(handler)

	// IP 1: Make 2 requests (limit reached)
	for i := 1; i <= 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("IP1 Request %d failed: expected 200, got %d", i, w.Code)
		}
	}

	// IP 2: Should be able to make requests normally
	for i := 1; i <= 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.2:1234"
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("IP2 Request %d failed: expected 200, got %d", i, w.Code)
		}
	}

	// IP 1: 3rd request should be blocked
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("IP1 3rd request failed: expected 429, got %d", w.Code)
	}

	// IP 2: 3rd request should also be blocked
	req = httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.2:1234"
	w = httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("IP2 3rd request failed: expected 429, got %d", w.Code)
	}
}
