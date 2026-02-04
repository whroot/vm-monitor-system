package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 基于内存的令牌桶限流器
type RateLimiter struct {
	rate       int           // 每秒生成的令牌数
	capacity   int           // 桶容量
	tokens     int           // 当前令牌数
	lastUpdate time.Time     // 上次更新时间
	mu         sync.Mutex    // 互斥锁
}

// NewRateLimiter 创建限流器
func NewRateLimiter(rate, capacity int) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		capacity:   capacity,
		tokens:     capacity,
		lastUpdate: time.Now(),
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	// 计算经过时间生成的令牌
	elapsed := now.Sub(rl.lastUpdate).Seconds()
	rl.tokens = min(rl.capacity, rl.tokens+int(elapsed*float64(rl.rate)))
	rl.lastUpdate = now

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// IPRateLimiter 基于IP的限流器
type IPRateLimiter struct {
	limiters map[string]*RateLimiter
	mu       sync.RWMutex
	rate     int
	capacity int
}

// NewIPRateLimiter 创建IP限流器
func NewIPRateLimiter(rate, capacity int) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*RateLimiter),
		rate:     rate,
		capacity: capacity,
	}
}

// GetLimiter 获取IP对应的限流器
func (irl *IPRateLimiter) GetLimiter(ip string) *RateLimiter {
	irl.mu.RLock()
	limiter, exists := irl.limiters[ip]
	irl.mu.RUnlock()

	if !exists {
		irl.mu.Lock()
		limiter = NewRateLimiter(irl.rate, irl.capacity)
		irl.limiters[ip] = limiter
		irl.mu.Unlock()
	}
	return limiter
}

// Cleanup 清理过期的限流器
func (irl *IPRateLimiter) Cleanup() {
	irl.mu.Lock()
	defer irl.mu.Unlock()

	// 这里可以添加清理逻辑，删除长时间未使用的IP限流器
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	LoginRate     int // 登录接口限流（次/分钟）
	GeneralRate   int // 普通接口限流（次/分钟）
	BurstCapacity int // 突发容量
}

// DefaultRateLimitConfig 默认配置
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		LoginRate:     5,   // 登录：5次/分钟
		GeneralRate:   100, // 普通：100次/分钟
		BurstCapacity: 10,  // 突发容量
	}
}

// RateLimit 限流中间件
func RateLimit(config RateLimitConfig) gin.HandlerFunc {
	// 创建IP限流器
	loginLimiter := NewIPRateLimiter(config.LoginRate, config.BurstCapacity)
	generalLimiter := NewIPRateLimiter(config.GeneralRate, config.BurstCapacity)

	return func(c *gin.Context) {
		// 获取客户端IP
		clientIP := c.ClientIP()

		// 根据路径选择限流器
		var limiter *RateLimiter
		if c.Request.URL.Path == "/api/v1/auth/login" {
			limiter = loginLimiter.GetLimiter(clientIP)
		} else {
			limiter = generalLimiter.GetLimiter(clientIP)
		}

		// 检查是否允许请求
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// StrictRateLimit 严格限流（用于敏感操作）
func StrictRateLimit(rate, capacity int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(rate, capacity)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		l := limiter.GetLimiter(clientIP)

		if !l.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求频率超限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
