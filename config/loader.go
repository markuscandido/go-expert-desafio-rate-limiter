package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func LoadConfig() *RateLimiterConfig {
	godotenv.Load()

	config := NewConfig()

	// Load IP-based limiting config
	if val := os.Getenv("RATE_LIMITER_ENABLE_IP"); val != "" {
		config.EnableIPLimit = val == "true"
	}
	if val := os.Getenv("RATE_LIMITER_MAX_REQUESTS_IP"); val != "" {
		if maxReq, err := strconv.Atoi(val); err == nil {
			config.MaxRequestsIP = maxReq
		}
	}
	if val := os.Getenv("RATE_LIMITER_BLOCK_DURATION_IP"); val != "" {
		if blockDur, err := strconv.Atoi(val); err == nil {
			config.BlockDurationIP = blockDur
		}
	}

	// Load Token-based limiting config
	if val := os.Getenv("RATE_LIMITER_ENABLE_TOKEN"); val != "" {
		config.EnableTokenLimit = val == "true"
	}
	if val := os.Getenv("RATE_LIMITER_MAX_REQUESTS_TOKEN"); val != "" {
		if maxReq, err := strconv.Atoi(val); err == nil {
			config.MaxRequestsToken = maxReq
		}
	}
	if val := os.Getenv("RATE_LIMITER_BLOCK_DURATION_TOKEN"); val != "" {
		if blockDur, err := strconv.Atoi(val); err == nil {
			config.BlockDurationToken = blockDur
		}
	}

	// Load Redis config
	if val := os.Getenv("REDIS_ADDR"); val != "" {
		config.RedisAddr = val
	}
	if val := os.Getenv("REDIS_DB"); val != "" {
		if db, err := strconv.Atoi(val); err == nil {
			config.RedisDB = db
		}
	}
	if val := os.Getenv("REDIS_PASS"); val != "" {
		config.RedisPass = val
	}

	return config
}
