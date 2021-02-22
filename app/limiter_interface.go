package main

import (
	"time"
)

// LimiterRepository defines the data access interface
type LimiterRepository interface {
	Limit(ipaddr string) (int64, time.Duration, error)
}
