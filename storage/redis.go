package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStrategy struct {
	client *redis.Client
}

func NewRedisStrategy(addr string, db int, password string) (*RedisStrategy, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       db,
		Password: password,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStrategy{client: client}, nil
}

func (r *RedisStrategy) CheckAndIncrement(ctx context.Context, key string, maxRequests int, windowSeconds int) (allowed bool, err error) {
	// Check if blocked
	isBlocked, err := r.IsBlocked(ctx, key)
	if err != nil {
		return false, err
	}
	if isBlocked {
		return false, nil
	}

	// Get current data
	data, err := r.GetData(ctx, key)
	if err != nil {
		return false, err
	}

	// Initialize if new
	if data == nil {
		data = &LimiterData{
			Count:     0,
			ExpiresAt: time.Now().Add(time.Duration(windowSeconds) * time.Second),
			IsBlocked: false,
		}
	}

	// Check if window expired
	if time.Now().After(data.ExpiresAt) {
		data.Count = 0
		data.ExpiresAt = time.Now().Add(time.Duration(windowSeconds) * time.Second)
	}

	// Increment counter
	data.Count++

	// Store updated data
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return false, err
	}

	duration := time.Until(data.ExpiresAt)
	if duration < 0 {
		duration = time.Duration(windowSeconds) * time.Second
	}

	err = r.client.Set(ctx, key, dataJSON, duration).Err()
	if err != nil {
		return false, err
	}

	// Check if limit exceeded
	return data.Count <= maxRequests, nil
}

func (r *RedisStrategy) IsBlocked(ctx context.Context, key string) (blocked bool, err error) {
	blockedKey := key + ":blocked"
	result, err := r.client.Get(ctx, blockedKey).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return result == "true", nil
}

func (r *RedisStrategy) Block(ctx context.Context, key string, durationSeconds int) error {
	blockedKey := key + ":blocked"
	duration := time.Duration(durationSeconds) * time.Second
	return r.client.Set(ctx, blockedKey, "true", duration).Err()
}

func (r *RedisStrategy) Reset(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	blockedKey := key + ":blocked"
	return r.client.Del(ctx, blockedKey).Err()
}

func (r *RedisStrategy) GetData(ctx context.Context, key string) (*LimiterData, error) {
	result, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var data LimiterData
	err = json.Unmarshal([]byte(result), &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (r *RedisStrategy) Close() error {
	return r.client.Close()
}
