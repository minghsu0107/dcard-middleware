package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// RateLimiterMiddleware implements rate limitation for all http endpoints
func RateLimiterMiddleware(limiterRepo LimiterRepository, maxVisitCount int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		var err error
		var exist bool
		exist, err = limiterRepo.Exists(ip)
		if err != nil {
			middlewareErrorLogging(err)
			c.Abort()
			return
		}

		var newVisitCount int64
		if !exist {
			if err := limiterRepo.SetVisitCount(ip, 1); err != nil {
				middlewareErrorLogging(err)
				c.Abort()
				return
			}
			newVisitCount = 1
		} else {
			var err error
			newVisitCount, err = limiterRepo.IncrVisitCountByIP(ip)
			if err != nil {
				middlewareErrorLogging(err)
				c.Abort()
				return
			}
		}

		var ttl time.Duration
		ttl, err = limiterRepo.GetTTL(ip)
		if err != nil {
			middlewareErrorLogging(err)
			c.Abort()
			return
		}

		remaining := maxVisitCount - newVisitCount
		if remaining < 0 {
			remaining = 0
		}
		c.Writer.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		c.Writer.Header().Set("X-RateLimit-Reset", strconv.Itoa(int(ttl.Seconds())))
		if newVisitCount > maxVisitCount {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		c.Next()
	}
}

func middlewareErrorLogging(err error) {
	log.WithFields(log.Fields{
		"type": "middleware",
	}).Error(err)
}
