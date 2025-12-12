package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"github.com/markuscandido/go-expert-desafio-rate-limiter/pkg/logger"
)

func LoadConfig() *RateLimiterConfig {
	godotenv.Load()

	config := NewConfig()
	logger.Debug("Loading configuration from environment")

	// Load IP-based limiting config
	if val := os.Getenv("RATE_LIMITER_ENABLE_IP"); val != "" {
		config.EnableIPLimit = val == "true"
		logger.Debug("Configuration loaded", "RATE_LIMITER_ENABLE_IP", config.EnableIPLimit)
	}
	if val := os.Getenv("RATE_LIMITER_MAX_REQUESTS_IP"); val != "" {
		if maxReq, err := strconv.Atoi(val); err == nil {
			config.MaxRequestsIP = maxReq
			logger.Debug("Configuration loaded", "RATE_LIMITER_MAX_REQUESTS_IP", maxReq)
		} else {
			logger.Warn("Invalid value for RATE_LIMITER_MAX_REQUESTS_IP", "value", val, "error", err)
		}
	}
	if val := os.Getenv("RATE_LIMITER_BLOCK_DURATION_IP"); val != "" {
		if blockDur, err := strconv.Atoi(val); err == nil {
			config.BlockDurationIP = blockDur
			logger.Debug("Configuration loaded", "RATE_LIMITER_BLOCK_DURATION_IP", blockDur)
		} else {
			logger.Warn("Invalid value for RATE_LIMITER_BLOCK_DURATION_IP", "value", val, "error", err)
		}
	}

	// Load Token-based limiting config
	if val := os.Getenv("RATE_LIMITER_ENABLE_TOKEN"); val != "" {
		config.EnableTokenLimit = val == "true"
		logger.Debug("Configuration loaded", "RATE_LIMITER_ENABLE_TOKEN", config.EnableTokenLimit)
	}
	if val := os.Getenv("RATE_LIMITER_MAX_REQUESTS_TOKEN"); val != "" {
		if maxReq, err := strconv.Atoi(val); err == nil {
			config.MaxRequestsToken = maxReq
			logger.Debug("Configuration loaded", "RATE_LIMITER_MAX_REQUESTS_TOKEN", maxReq)
		} else {
			logger.Warn("Invalid value for RATE_LIMITER_MAX_REQUESTS_TOKEN", "value", val, "error", err)
		}
	}
	if val := os.Getenv("RATE_LIMITER_BLOCK_DURATION_TOKEN"); val != "" {
		if blockDur, err := strconv.Atoi(val); err == nil {
			config.BlockDurationToken = blockDur
			logger.Debug("Configuration loaded", "RATE_LIMITER_BLOCK_DURATION_TOKEN", blockDur)
		} else {
			logger.Warn("Invalid value for RATE_LIMITER_BLOCK_DURATION_TOKEN", "value", val, "error", err)
		}
	}

	// Load Redis config
	if val := os.Getenv("REDIS_ADDR"); val != "" {
		config.RedisAddr = val
		logger.Debug("Configuration loaded", "REDIS_ADDR", val)
	}
	if val := os.Getenv("REDIS_DB"); val != "" {
		if db, err := strconv.Atoi(val); err == nil {
			config.RedisDB = db
			logger.Debug("Configuration loaded", "REDIS_DB", db)
		} else {
			logger.Warn("Invalid value for REDIS_DB", "value", val, "error", err)
		}
	}
	if val := os.Getenv("REDIS_PASS"); val != "" {
		config.RedisPass = val
		logger.Debug("Configuration loaded", "REDIS_PASS", "***")
	}

	logger.Info("Configuration loaded successfully",
		"ipLimitEnabled", config.EnableIPLimit,
		"tokenLimitEnabled", config.EnableTokenLimit,
	)
	return config
}
