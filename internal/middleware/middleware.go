package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/markuscandido/go-expert-desafio-rate-limiter/internal/limiter"
	"github.com/markuscandido/go-expert-desafio-rate-limiter/pkg/logger"
)

type RateLimiterMiddleware struct {
	limiter *limiter.RateLimiter
}

func NewRateLimiterMiddleware(l *limiter.RateLimiter) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		limiter: l,
	}
}

const ErrorMessage = "you have reached the maximum number of requests or actions allowed within a certain time frame"

func (m *RateLimiterMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		token := getToken(r)

		allowed, blockDuration, err := m.limiter.AllowRequest(r.Context(), ip, token)

		if err != nil {
			logger.Error("Rate limiter error",
				"path", r.RequestURI,
				"ip", ip,
				"hasToken", token != "",
				"error", err,
			)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !allowed {
			logger.Warn("Rate limit exceeded",
				"path", r.RequestURI,
				"ip", ip,
				"hasToken", token != "",
				"blockDuration", blockDuration,
			)
			w.Header().Set("Retry-After", strconv.Itoa(blockDuration))
			http.Error(w, ErrorMessage, http.StatusTooManyRequests)
			return
		}

		logger.Debug("Request allowed",
			"path", r.RequestURI,
			"method", r.Method,
			"ip", ip,
			"hasToken", token != "",
		)
		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

func getToken(r *http.Request) string {
	return r.Header.Get("API_KEY")
}
