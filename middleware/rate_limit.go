package middleware

import (
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"enterprise-agent/backend/models"
)

type ipRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	limit    rate.Limit
	burst    int
}

func newIPRateLimiter(rps float64, burst int) *ipRateLimiter {
	if rps <= 0 {
		rps = 1
	}
	if burst <= 0 {
		burst = 5
	}
	return &ipRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		limit:    rate.Limit(rps),
		burst:    burst,
	}
}

func (l *ipRateLimiter) getLimiter(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.limiters[ip]
	if !ok {
		entry = rate.NewLimiter(l.limit, l.burst)
		l.limiters[ip] = entry
	}
	return entry
}

// RateLimitByIP 按客户端 IP 限流
func RateLimitByIP(requestsPerMinute int, burst int) gin.HandlerFunc {
	if requestsPerMinute <= 0 {
		requestsPerMinute = 30
	}
	rps := float64(requestsPerMinute) / 60.0
	limiter := newIPRateLimiter(rps, burst)

	return func(c *gin.Context) {
		if !limiter.getLimiter(c.ClientIP()).Allow() {
			models.SendError(c, 429, 2001, "Too many requests, please try again later")
			c.Abort()
			return
		}
		c.Next()
	}
}

// AuthRateLimit 登录/刷新等认证接口限流（约 12 次/分钟，突发 6 次）
func AuthRateLimit() gin.HandlerFunc {
	return RateLimitByIP(12, 6)
}

// DeviceRegisterRateLimit 设备注册限流（约 6 次/分钟，突发 3 次）
func DeviceRegisterRateLimit() gin.HandlerFunc {
	return RateLimitByIP(6, 3)
}
