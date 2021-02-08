package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// GinMiddleware defines the middleware interface
type GinMiddleware interface {
	Provide() gin.HandlerFunc
}

// GinMiddlewareCollection is a list of GinMiddlewares
// it is a helper for dependency injection
type GinMiddlewareCollection []GinMiddleware

// RateLimiterMiddleware is the type of rate limiter middleware
type RateLimiterMiddleware struct {
	Repo          LimiterRepository
	MaxVisitCount int64
	logger        *log.Entry
}

// NewGinMiddlewareCollection is the factory of GinMiddlewareCollection
func NewGinMiddlewareCollection(ginMiddlewares ...GinMiddleware) *GinMiddlewareCollection {
	var ginMiddlewareCollection GinMiddlewareCollection
	for _, ginMiddleware := range ginMiddlewares {
		ginMiddlewareCollection = append(ginMiddlewareCollection, ginMiddleware)
	}
	return &ginMiddlewareCollection
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

// Provide method is the implementation of the middleware inteface
func (m *RateLimiterMiddleware) Provide() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		var err error
		var exist bool
		exist, err = m.Repo.Exists(ip)
		if err != nil {
			m.logger.Error(err)
			c.Abort()
			return
		}

		var newVisitCount int64
		if !exist {
			if err := m.Repo.SetVisitCount(ip, 1); err != nil {
				m.logger.Error(err)
				c.Abort()
				return
			}
			newVisitCount = 1
		} else {
			var err error
			newVisitCount, err = m.Repo.IncrVisitCountByIP(ip)
			if err != nil {
				m.logger.Error(err)
				c.Abort()
				return
			}
		}

		var ttl time.Duration
		ttl, err = m.Repo.GetTTL(ip)
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
