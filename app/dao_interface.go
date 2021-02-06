package main

import (
	"time"
)

// LimiterRepository defines the data access interface
type LimiterRepository interface {
	Exists(ipaddr string) (bool, error)
	GetTTL(ipaddr string) (time.Duration, error)
	SetVisitCount(ipaddr string, count int) error
	IncrVisitCountByIP(ipaddr string) (int64, error)
}
