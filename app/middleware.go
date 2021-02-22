package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// RateLimiterMiddleware is the type of rate limiter middleware
type RateLimiterMiddleware struct {
	Repo          LimiterRepository
	MaxVisitCount int64
	logger        *log.Entry
}

// NewRateLimiterMiddleware is the factory of RateLimiterMiddleware
func NewRateLimiterMiddleware(config *Config, repo LimiterRepository) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		Repo:          repo,
		MaxVisitCount: config.MaxVisitCount,
		logger: config.Logger.ContextLogger.WithFields(log.Fields{
			"type": "middleware:rateLimiter",
		}),
	}
}

// Provide method returns a gin handler function
func (m *RateLimiterMiddleware) Provide() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		m.logger.Infof("client ip: %v\n", ip)

		newVisitCount, ttl, err := m.Repo.Limit(ip)
		if err != nil {
			m.logger.Error(err)
			c.Abort()
			return
		}

		remaining := m.MaxVisitCount - newVisitCount
		if remaining < 0 {
			remaining = 0
		}
		c.Writer.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		c.Writer.Header().Set("X-RateLimit-Reset", strconv.Itoa(int(ttl.Seconds())))
		if newVisitCount > m.MaxVisitCount {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		c.Next()
	}
}
