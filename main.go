package main

import (
	"context"
	"net/http"

	"github.com/markuscandido/go-expert-desafio-rate-limiter/config"
	"github.com/markuscandido/go-expert-desafio-rate-limiter/limiter"
	"github.com/markuscandido/go-expert-desafio-rate-limiter/middleware"
	"github.com/markuscandido/go-expert-desafio-rate-limiter/pkg/logger"
	"github.com/markuscandido/go-expert-desafio-rate-limiter/storage"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	logger.Info("Starting rate limiter server",
		"maxRequestsIP", cfg.MaxRequestsIP,
		"enableIPLimit", cfg.EnableIPLimit,
		"maxRequestsToken", cfg.MaxRequestsToken,
		"enableTokenLimit", cfg.EnableTokenLimit,
		"redisAddr", cfg.RedisAddr,
	)

	// Initialize Redis storage
	redisStrategy, err := storage.NewRedisStrategy(cfg.RedisAddr, cfg.RedisDB, cfg.RedisPass)
	if err != nil {
		logger.Fatal("Failed to initialize Redis", "error", err)
	}
	defer redisStrategy.Close()

	// Create rate limiter
	rateLimiter := limiter.NewRateLimiter(redisStrategy, cfg)

	// Create middleware
	rateLimiterMiddleware := middleware.NewRateLimiterMiddleware(rateLimiter)

	// Create a simple handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Hello from rate-limiter server!"}`))
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy"}`))
	})

	// Wrap with rate limiter middleware
	handler := rateLimiterMiddleware.Handler(mux)

	// Start server
	addr := ":8080"
	logger.Info("Server listening", "address", addr)
	if err := http.ListenAndServe(addr, handler); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Server error", "error", err)
	}
}

// Graceful shutdown can be added here
func shutdown(ctx context.Context, server *http.Server) {
	server.Shutdown(ctx)
}
