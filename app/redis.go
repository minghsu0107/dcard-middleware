package main

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisLimiterRepository implements LimiterRepository interface
type RedisLimiterRepository struct {
	client     *redis.Client
	ctx        context.Context
	expiration time.Duration
}

// NewRedisLimiterRepository is the factory of RedisLimiterRepository
func NewRedisLimiterRepository(config *Config) (LimiterRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:        config.RedisConfig.Addr,
		Password:    config.RedisConfig.Password,
		PoolSize:    config.RedisConfig.PoolSize,
		MaxRetries:  config.RedisConfig.MaxRetries,
		IdleTimeout: config.RedisConfig.IdleTimeout,
	})
	ctx := context.Background()
	pong, err := client.Ping(ctx).Result()
	if err == redis.Nil || err != nil {
		return nil, err
	}
	config.Logger.ContextLogger.WithField("type", "setup:redis").Info("successful Redis Connection: " + pong)
	return &RedisLimiterRepository{
		client:     client,
		ctx:        ctx,
		expiration: config.RedisConfig.CacheExpiration,
	}, nil
}

// GetVisitCount checks the existence of the given ipaddr
// and returns true and the related visit count and ttl if the entry exists
// otherwise returns false
func (r *RedisLimiterRepository) GetVisitCount(ipaddr string) (*Record, bool, error) {
	var err error
	var count int
	count, err = r.client.Get(r.ctx, ipaddr).Int()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var ttl time.Duration
	ttl, err = r.client.TTL(r.ctx, ipaddr).Result()
	if err != nil {
		return nil, true, err
	}

	return &Record{
		Count: count,
		TTL:   ttl,
	}, true, nil
}

// SetVisitCount sets visit count of the given ip
func (r *RedisLimiterRepository) SetVisitCount(ipaddr string, count int) error {
	if err := r.client.Set(r.ctx, ipaddr, count, r.expiration).Err(); err != nil {
		return err
	}
	return nil
}

// IncrVisitCountByIP increments the visit count of the given ip by one
func (r *RedisLimiterRepository) IncrVisitCountByIP(ipaddr string) error {
	if _, err := r.client.Incr(r.ctx, ipaddr).Result(); err != nil {
		return err
	}
	return nil
}
