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

// GetTTL returns the time-to-expire of the given ip
func (r *RedisLimiterRepository) GetTTL(ipaddr string) (time.Duration, error) {
	var ttl time.Duration
	var err error
	ttl, err = r.client.TTL(r.ctx, ipaddr).Result()
	if err != nil {
		return 0, err
	}

	return ttl, nil
}

// IncrVisitCountByIP increments the visit count of the given ip by one
func (r *RedisLimiterRepository) IncrVisitCountByIP(ipaddr string) (int64, error) {
	var newCount int64
	var err error
	newCount, err = r.client.Incr(r.ctx, ipaddr).Result()
	if err != nil {
		return -1, err
	}
	return newCount, nil
}

// SetVisitCount sets visit count of the given ip with ttl if the key does not exist, otherwise do nothing
func (r *RedisLimiterRepository) SetVisitCount(ipaddr string, count int) error {
	if err := r.client.SetNX(r.ctx, ipaddr, count, r.expiration).Err(); err != nil {
		return err
	}
	return nil
}

// Exists check whether the key exists
func (r *RedisLimiterRepository) Exists(ipaddr string) (bool, error) {
	var err error
	intRes, err := r.client.Exists(r.ctx, ipaddr).Result()
	if err != nil {
		return false, err
	}
	exist := (intRes != 0)
	return exist, nil
}
