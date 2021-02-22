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

// Limit exeucutes SetVisitCountNX, IncrVisitCountByIP, and GetTTL in a pipeline
// to reduce networking overhead. It returns the updated visit count and current TTL of the given ip
func (r *RedisLimiterRepository) Limit(ipaddr string) (int64, time.Duration, error) {
	pipe := r.client.Pipeline()
	pipedCmds := []interface{}{
		pipe.SetNX(r.ctx, ipaddr, 0, r.expiration),
		pipe.Incr(r.ctx, ipaddr),
		pipe.TTL(r.ctx, ipaddr),
	}
	_, err := pipe.Exec(r.ctx)
	if err != nil {
		return 0, 0, err
	}

	executedSetVisitCountNX := pipedCmds[0].(*redis.BoolCmd)
	executedIncrVisitCountByIP := pipedCmds[1].(*redis.IntCmd)
	executedGetTTL := pipedCmds[2].(*redis.DurationCmd)

	var newCount int64
	var ttl time.Duration

	if err = executedSetVisitCountNX.Err(); err != nil {
		return 0, 0, err
	}
	if newCount, err = executedIncrVisitCountByIP.Result(); err != nil {
		return 0, 0, err
	}
	if ttl, err = executedGetTTL.Result(); err != nil {
		return 0, 0, err
	}
	return newCount, ttl, nil
}
