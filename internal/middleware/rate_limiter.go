package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// 简单的内存限流器
type RateLimiter struct {
	clients map[string]*ClientInfo
	mutex   sync.RWMutex
	rate    int           // 每秒允许的请求数
	window  time.Duration // 时间窗口
}

type ClientInfo struct {
	requests  int
	lastReset time.Time
}

func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*ClientInfo),
		rate:    rate,
		window:  window,
	}
}

func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	client, exists := rl.clients[clientIP]

	if !exists {
		rl.clients[clientIP] = &ClientInfo{
			requests:  1,
			lastReset: now,
		}
		return true
	}

	// 如果超过时间窗口，重置计数
	if now.Sub(client.lastReset) > rl.window {
		client.requests = 1
		client.lastReset = now
		return true
	}

	// 检查是否超过限制
	if client.requests >= rl.rate {
		return false
	}

	client.requests++
	return true
}

func RateLimiterMiddleware() gin.HandlerFunc {
	limiter := NewRateLimiter(100, time.Minute) // 每分钟100个请求

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !limiter.Allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "请求过于频繁",
				"message": "请稍后重试",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
