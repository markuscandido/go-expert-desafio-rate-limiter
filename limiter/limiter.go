package limiter

import (
	"context"
	"fmt"
	"strings"

	"github.com/markuscandido/go-expert-desafio-rate-limiter/config"
	"github.com/markuscandido/go-expert-desafio-rate-limiter/pkg/logger"
	"github.com/markuscandido/go-expert-desafio-rate-limiter/storage"
)

type RateLimiter struct {
	storage storage.Strategy
	config  *config.RateLimiterConfig
}

func NewRateLimiter(st storage.Strategy, cfg *config.RateLimiterConfig) *RateLimiter {
	return &RateLimiter{
		storage: st,
		config:  cfg,
	}
}

// AllowRequest checks if a request should be allowed based on IP and/or token
// Returns (allowed, blockDuration, error)
func (rl *RateLimiter) AllowRequest(ctx context.Context, ip string, token string) (allowed bool, blockDuration int, err error) {
	var limit *RequestLimit

	// Check token limit first (takes precedence over IP limit)
	if rl.config.EnableTokenLimit && token != "" {
		token = strings.TrimSpace(token)
		limit, err = rl.checkTokenLimit(ctx, token)
		if err != nil {
			return false, 0, err
		}
		if limit != nil {
			return limit.Allowed, limit.BlockDuration, nil
		}
	}

	// Check IP limit
	if rl.config.EnableIPLimit && ip != "" {
		limit, err = rl.checkIPLimit(ctx, ip)
		if err != nil {
			return false, 0, err
		}
		if limit != nil {
			return limit.Allowed, limit.BlockDuration, nil
		}
	}

	// If no limit is enabled, allow the request
	return true, 0, nil
}

func (rl *RateLimiter) checkIPLimit(ctx context.Context, ip string) (*RequestLimit, error) {
	if !rl.config.EnableIPLimit {
		return nil, nil
	}

	key := fmt.Sprintf("ip:%s", ip)

	// Check if blocked
	isBlocked, err := rl.storage.IsBlocked(ctx, key)
	if err != nil {
		logger.Error("Failed to check if IP is blocked",
			"ip", ip,
			"error", err,
		)
		return nil, err
	}

	if isBlocked {
		logger.Warn("IP blocked",
			"ip", ip,
			"blockDuration", rl.config.BlockDurationIP,
		)
		return &RequestLimit{
			Allowed:       false,
			BlockDuration: rl.config.BlockDurationIP,
		}, nil
	}

	// Check and increment
	allowed, err := rl.storage.CheckAndIncrement(ctx, key, rl.config.MaxRequestsIP, 1)
	if err != nil {
		logger.Error("Failed to check and increment IP limit",
			"ip", ip,
			"error", err,
		)
		return nil, err
	}

	if !allowed {
		// Block the IP
		err = rl.storage.Block(ctx, key, rl.config.BlockDurationIP)
		if err != nil {
			logger.Error("Failed to block IP",
				"ip", ip,
				"blockDuration", rl.config.BlockDurationIP,
				"error", err,
			)
			return nil, err
		}
		logger.Warn("IP rate limit exceeded",
			"ip", ip,
			"blockDuration", rl.config.BlockDurationIP,
		)
		return &RequestLimit{
			Allowed:       false,
			BlockDuration: rl.config.BlockDurationIP,
		}, nil
	}

	return &RequestLimit{
		Allowed:       true,
		BlockDuration: 0,
	}, nil
}

func (rl *RateLimiter) checkTokenLimit(ctx context.Context, token string) (*RequestLimit, error) {
	if !rl.config.EnableTokenLimit {
		return nil, nil
	}

	key := fmt.Sprintf("token:%s", token)

	// Check if blocked
	isBlocked, err := rl.storage.IsBlocked(ctx, key)
	if err != nil {
		logger.Error("Failed to check if token is blocked",
			"error", err,
		)
		return nil, err
	}

	if isBlocked {
		logger.Warn("Token blocked",
			"blockDuration", rl.config.BlockDurationToken,
		)
		return &RequestLimit{
			Allowed:       false,
			BlockDuration: rl.config.BlockDurationToken,
		}, nil
	}

	// Check and increment
	allowed, err := rl.storage.CheckAndIncrement(ctx, key, rl.config.MaxRequestsToken, 1)
	if err != nil {
		logger.Error("Failed to check and increment token limit",
			"error", err,
		)
		return nil, err
	}

	if !allowed {
		// Block the token
		err = rl.storage.Block(ctx, key, rl.config.BlockDurationToken)
		if err != nil {
			logger.Error("Failed to block token",
				"blockDuration", rl.config.BlockDurationToken,
				"error", err,
			)
			return nil, err
		}
		logger.Warn("Token rate limit exceeded",
			"blockDuration", rl.config.BlockDurationToken,
		)
		return &RequestLimit{
			Allowed:       false,
			BlockDuration: rl.config.BlockDurationToken,
		}, nil
	}

	return &RequestLimit{
		Allowed:       true,
		BlockDuration: 0,
	}, nil
}

// RequestLimit represents the result of a rate limit check
type RequestLimit struct {
	Allowed       bool
	BlockDuration int
}
