package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/markuscandido/go-expert-desafio-rate-limiter/limiter"
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
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !allowed {
			w.Header().Set("Retry-After", string(rune(blockDuration)))
			http.Error(w, ErrorMessage, http.StatusTooManyRequests)
			return
		}

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
