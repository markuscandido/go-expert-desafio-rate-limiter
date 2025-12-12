package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/markuscandido/go-expert-desafio-rate-limiter/internal/config"
	"github.com/markuscandido/go-expert-desafio-rate-limiter/internal/limiter"
	"github.com/markuscandido/go-expert-desafio-rate-limiter/internal/storage"
)

type MockStorageForMiddleware struct {
	counter map[string]int
	blocked map[string]bool
}

func NewMockStorageForMiddleware() *MockStorageForMiddleware {
	return &MockStorageForMiddleware{
		counter: make(map[string]int),
		blocked: make(map[string]bool),
	}
}

func (m *MockStorageForMiddleware) CheckAndIncrement(ctx context.Context, key string, maxRequests int, windowSeconds int) (bool, error) {
	if m.blocked[key] {
		return false, nil
	}
	m.counter[key]++
	return m.counter[key] <= maxRequests, nil
}

func (m *MockStorageForMiddleware) IsBlocked(ctx context.Context, key string) (bool, error) {
	return m.blocked[key], nil
}

func (m *MockStorageForMiddleware) Block(ctx context.Context, key string, durationSeconds int) error {
	m.blocked[key] = true
	return nil
}

func (m *MockStorageForMiddleware) Reset(ctx context.Context, key string) error {
	delete(m.counter, key)
	delete(m.blocked, key)
	return nil
}

func (m *MockStorageForMiddleware) GetData(ctx context.Context, key string) (*storage.LimiterData, error) {
	return nil, nil
}

func (m *MockStorageForMiddleware) Close() error {
	return nil
}

func TestMiddlewareBlocksExceededRequests(t *testing.T) {
	mockStorage := NewMockStorageForMiddleware()
	cfg := &config.RateLimiterConfig{
		MaxRequestsIP:    2,
		BlockDurationIP:  60,
		EnableIPLimit:    true,
		EnableTokenLimit: false,
	}

	rateLimiter := limiter.NewRateLimiter(mockStorage, cfg)
	m := NewRateLimiterMiddleware(rateLimiter)

	// Mock handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	wrappedHandler := m.Handler(handler)

	// First 2 requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d should return 200, got %d", i+1, w.Code)
		}
	}

	// 3rd request should be blocked
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("3rd request should return 429, got %d", w.Code)
	}

	body := strings.TrimSpace(w.Body.String())
	if body != ErrorMessage {
		t.Errorf("Expected error message, got: %s", body)
	}
}

func TestMiddlewareWithToken(t *testing.T) {
	mockStorage := NewMockStorageForMiddleware()
	cfg := &config.RateLimiterConfig{
		MaxRequestsToken:   3,
		BlockDurationToken: 60,
		EnableIPLimit:      false,
		EnableTokenLimit:   true,
	}

	rateLimiter := limiter.NewRateLimiter(mockStorage, cfg)
	m := NewRateLimiterMiddleware(rateLimiter)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	wrappedHandler := m.Handler(handler)

	// 3 requests with token should succeed
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("API_KEY", "test-token")
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d should return 200, got %d", i+1, w.Code)
		}
	}

	// 4th request should be blocked
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("API_KEY", "test-token")
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("4th request should return 429, got %d", w.Code)
	}
}
