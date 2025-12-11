package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	fmt.Printf("Starting server with config:\n")
	fmt.Printf("  IP Limit: %d req/s (enabled: %v)\n", cfg.MaxRequestsIP, cfg.EnableIPLimit)
	fmt.Printf("  Token Limit: %d req/s (enabled: %v)\n", cfg.MaxRequestsToken, cfg.EnableTokenLimit)
	fmt.Printf("  Redis: %s\n", cfg.RedisAddr)

	// Initialize Redis storage
	redisStrategy, err := storage.NewRedisStrategy(cfg.RedisAddr, cfg.RedisDB, cfg.RedisPass)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
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
		w.Write([]byte(`{"message": "Hello from github.com/markuscandido/go-expert-desafio-rate-limiter server!"}`))
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
	fmt.Printf("Server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, handler); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

// Graceful shutdown can be added here
func shutdown(ctx context.Context, server *http.Server) {
	server.Shutdown(ctx)
}
