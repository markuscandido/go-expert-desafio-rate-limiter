package main

type RateLimiterConfig struct {
	// IP-based rate limiting
	MaxRequestsIP    int // Maximum requests per second from a single IP
	BlockDurationIP  int // Block duration in seconds for IP
	EnableIPLimit    bool

	// Token-based rate limiting
	MaxRequestsToken   int // Maximum requests per second for a token
	BlockDurationToken int // Block duration in seconds for token
	EnableTokenLimit   bool

	// Redis configuration
	RedisAddr string
	RedisDB   int
	RedisPass string
}

func NewConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		MaxRequestsIP:      10,
		BlockDurationIP:    60,
		EnableIPLimit:      true,
		MaxRequestsToken:   100,
		BlockDurationToken: 60,
		EnableTokenLimit:   true,
		RedisAddr:          "localhost:6379",
		RedisDB:            0,
		RedisPass:          "",
	}
}
