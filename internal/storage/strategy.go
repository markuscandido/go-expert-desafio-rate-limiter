package storage

import (
	"context"
	"time"
)

// LimiterData represents the rate limiter data stored in Redis
type LimiterData struct {
	Count     int       `json:"count"`
	ExpiresAt time.Time `json:"expires_at"`
	IsBlocked bool      `json:"is_blocked"`
}

// Strategy defines the interface for rate limiter storage
type Strategy interface {
	// CheckAndIncrement checks if the request is allowed and increments the counter
	CheckAndIncrement(ctx context.Context, key string, maxRequests int, windowSeconds int) (allowed bool, err error)

	// IsBlocked checks if a key is blocked
	IsBlocked(ctx context.Context, key string) (blocked bool, err error)

	// Block blocks a key for the specified duration
	Block(ctx context.Context, key string, durationSeconds int) error

	// Reset resets the counter for a key
	Reset(ctx context.Context, key string) error

	// GetData retrieves the current data for a key
	GetData(ctx context.Context, key string) (*LimiterData, error)

	// Close closes the storage connection
	Close() error
}
