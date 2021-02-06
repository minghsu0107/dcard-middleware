package main

import (
	"time"
)

// LimiterRepository defines the data access interface
type LimiterRepository interface {
	GetVisitCount(ipaddr string) (*Record, bool, error)
	SetVisitCount(ipaddr string, count int) error
	IncrVisitCountByIP(ipaddr string) error
}

// Record contains client visiting count and ttl
type Record struct {
	Count int
	TTL   time.Duration
}
