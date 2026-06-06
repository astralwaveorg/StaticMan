package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// TokenBucket 令牌桶限流器
type TokenBucket struct {
	tokens     float64
	lastUpdate time.Time
	capacity   float64
	rate       float64 // tokens per second
}

// Allow 检查是否允许通过
func (tb *TokenBucket) Allow() bool {
	now := time.Now()
	elapsed := now.Sub(tb.lastUpdate).Seconds()
	tb.tokens = min(tb.tokens+elapsed*tb.rate, tb.capacity)
	tb.lastUpdate = now
	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

// RateLimiter 限流器管理
type RateLimiter struct {
	mu       sync.RWMutex
	buckets  map[string]*TokenBucket
	capacity float64
	rate     float64 // tokens per second
}

// NewRateLimiter 创建限流器
func NewRateLimiter(capacity, rate float64) *RateLimiter {
	return &RateLimiter{
		buckets:  make(map[string]*TokenBucket),
		capacity: capacity,
		rate:     rate,
	}
}

// Allow 检查客户端 IP 是否允许通过
func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.mu.Lock()
	bucket, ok := rl.buckets[clientIP]
	if !ok {
		bucket = &TokenBucket{
			tokens:     rl.capacity,
			lastUpdate: time.Now(),
			capacity:   rl.capacity,
			rate:       rl.rate,
		}
		rl.buckets[clientIP] = bucket
	}
	rl.mu.Unlock()
	return bucket.Allow()
}

// Cleanup 清理长时间未使用的桶（每 10 分钟调用一次）
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	now := time.Now()
	for ip, bucket := range rl.buckets {
		if now.Sub(bucket.lastUpdate) > 10*time.Minute {
			delete(rl.buckets, ip)
		}
	}
	rl.mu.Unlock()
}

// getClientIP 获取客户端真实 IP
func getClientIP(r *http.Request) string {
	// 优先从 X-Forwarded-For 获取（Nginx 代理场景）
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return xff
	}
	xri := r.Header.Get("X-Real-Ip")
	if xri != "" {
		return xri
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

// RateLimitMiddleware 返回限流中间件
// limits 是路径前缀到 (容量, 速率) 的映射
func RateLimitMiddleware(limits map[string]*RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)
			for prefix, limiter := range limits {
				if len(r.URL.Path) >= len(prefix) && r.URL.Path[:len(prefix)] == prefix {
					if !limiter.Allow(clientIP) {
						w.Header().Set("Content-Type", "application/json")
						w.Header().Set("Retry-After", "60")
						w.WriteHeader(http.StatusTooManyRequests)
						w.Write([]byte(`{"error":"请求过于频繁，请稍后再试"}`))
						return
					}
					break
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
