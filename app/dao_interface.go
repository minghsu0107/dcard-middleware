package main

import (
	"time"
)

// LimiterRepository defines the data access interface
type LimiterRepository interface {
	GetTTL(ipaddr string) (time.Duration, error)
	IncrVisitCountByIP(ipaddr string) (int64, error)
	SetVisitCountNX(ipaddr string, count int) error
	// Exists(ipaddr string) (bool, error)
}
