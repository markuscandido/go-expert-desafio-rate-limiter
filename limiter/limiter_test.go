package limiter

import (
	"context"
	"testing"

	"github.com/markuscandido/go-expert-desafio-rate-limiter/config"
	"github.com/markuscandido/go-expert-desafio-rate-limiter/storage"
)

// MockStrategy is a mock implementation of storage.Strategy for testing
type MockStrategy struct {
	data    map[string]*storage.LimiterData
	blocked map[string]bool
}

func NewMockStrategy() *MockStrategy {
	return &MockStrategy{
		data:    make(map[string]*storage.LimiterData),
		blocked: make(map[string]bool),
	}
}

func (m *MockStrategy) CheckAndIncrement(ctx context.Context, key string, maxRequests int, windowSeconds int) (allowed bool, err error) {
	if m.data[key] == nil {
		m.data[key] = &storage.LimiterData{Count: 0}
	}

	m.data[key].Count++
	return m.data[key].Count <= maxRequests, nil
}

func (m *MockStrategy) IsBlocked(ctx context.Context, key string) (blocked bool, err error) {
	return m.blocked[key], nil
}

func (m *MockStrategy) Block(ctx context.Context, key string, durationSeconds int) error {
	m.blocked[key] = true
	return nil
}

func (m *MockStrategy) Reset(ctx context.Context, key string) error {
	delete(m.data, key)
	delete(m.blocked, key)
	return nil
}

func (m *MockStrategy) GetData(ctx context.Context, key string) (*storage.LimiterData, error) {
	return m.data[key], nil
}

func (m *MockStrategy) Close() error {
	return nil
}

func TestIPRateLimiting(t *testing.T) {
	mockStorage := NewMockStrategy()
	cfg := &config.RateLimiterConfig{
		MaxRequestsIP:    5,
		BlockDurationIP:  60,
		EnableIPLimit:    true,
		EnableTokenLimit: false,
	}

	rateLimiter := NewRateLimiter(mockStorage, cfg)
	ctx := context.Background()

	// First 5 requests should be allowed
	for i := 0; i < 5; i++ {
		allowed, _, err := rateLimiter.AllowRequest(ctx, "192.168.1.1", "")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 6th request should be blocked
	allowed, blockDuration, err := rateLimiter.AllowRequest(ctx, "192.168.1.1", "")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if allowed {
		t.Error("6th request should be blocked")
	}
	if blockDuration != 60 {
		t.Errorf("Expected block duration 60, got %d", blockDuration)
	}

	// Different IP should be allowed
	allowed, _, err = rateLimiter.AllowRequest(ctx, "192.168.1.2", "")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !allowed {
		t.Error("Request from different IP should be allowed")
	}
}

func TestTokenRateLimiting(t *testing.T) {
	mockStorage := NewMockStrategy()
	cfg := &config.RateLimiterConfig{
		MaxRequestsToken:   10,
		BlockDurationToken: 60,
		EnableIPLimit:      false,
		EnableTokenLimit:   true,
	}

	rateLimiter := NewRateLimiter(mockStorage, cfg)
	ctx := context.Background()

	// First 10 requests should be allowed
	for i := 0; i < 10; i++ {
		allowed, _, err := rateLimiter.AllowRequest(ctx, "", "token123")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 11th request should be blocked
	allowed, blockDuration, err := rateLimiter.AllowRequest(ctx, "", "token123")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if allowed {
		t.Error("11th request should be blocked")
	}
	if blockDuration != 60 {
		t.Errorf("Expected block duration 60, got %d", blockDuration)
	}

	// Different token should be allowed
	allowed, _, err = rateLimiter.AllowRequest(ctx, "", "token456")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !allowed {
		t.Error("Request with different token should be allowed")
	}
}

func TestTokenPrecedenceOverIP(t *testing.T) {
	mockStorage := NewMockStrategy()
	cfg := &config.RateLimiterConfig{
		MaxRequestsIP:      5,
		BlockDurationIP:    60,
		EnableIPLimit:      true,
		MaxRequestsToken:   100,
		BlockDurationToken: 60,
		EnableTokenLimit:   true,
	}

	rateLimiter := NewRateLimiter(mockStorage, cfg)
	ctx := context.Background()

	ip := "192.168.1.1"
	token := "premium_token"

	// Use up IP limit (5 requests)
	for i := 0; i < 5; i++ {
		allowed, _, _ := rateLimiter.AllowRequest(ctx, ip, "")
		if !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 6th request from same IP should be blocked
	allowed, _, _ := rateLimiter.AllowRequest(ctx, ip, "")
	if allowed {
		t.Error("6th request without token should be blocked")
	}

	// But with token, it should be allowed (token has higher limit)
	mockStorage.Reset(ctx, "ip:"+ip)
	allowed, _, _ = rateLimiter.AllowRequest(ctx, ip, token)
	if !allowed {
		t.Error("Request with token should be allowed (token precedence)")
	}
}

func TestDisabledLimits(t *testing.T) {
	mockStorage := NewMockStrategy()
	cfg := &config.RateLimiterConfig{
		EnableIPLimit:    false,
		EnableTokenLimit: false,
	}

	rateLimiter := NewRateLimiter(mockStorage, cfg)
	ctx := context.Background()

	// All requests should be allowed
	for i := 0; i < 10; i++ {
		allowed, _, err := rateLimiter.AllowRequest(ctx, "192.168.1.1", "")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !allowed {
			t.Errorf("Request %d should be allowed when limits are disabled", i+1)
		}
	}
}
