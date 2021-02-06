package main

import (
	"os"
	"strconv"
	"time"
)

// GetEnvWithDefault is a helper function for specifying a default env value
func GetEnvWithDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Config is a type for general configuration
type Config struct {
	Port        string
	GinMode     string
	Logger      *Logger
	RedisConfig *RedisConfig
}

// RedisConfig is the configuration for redis
type RedisConfig struct {
	Addr            string
	Password        string
	DB              int
	PoolSize        int
	MaxRetries      int
	IdleTimeout     time.Duration
	CacheExpiration time.Duration
}

// NewConfig is a factory for Config instance
func NewConfig() (*Config, error) {
	redisDB, err := strconv.Atoi(GetEnvWithDefault("REDIS_DB", "0"))
	redisPoolSize, err := strconv.Atoi(GetEnvWithDefault("REDIS_POOL_SIZE", "10"))
	if err != nil {
		return &Config{}, err
	}
	redisMaxRetries, err := strconv.Atoi(GetEnvWithDefault("REDIS_MAX_RETRIES", "3"))
	if err != nil {
		return &Config{}, err
	}
	redisIdleTimeout, err := strconv.ParseInt(GetEnvWithDefault("REDIS_IDLE_TIMEOUT", "60"), 10, 64)
	if err != nil {
		return &Config{}, err
	}
	cacheExpiration, err := strconv.ParseInt(GetEnvWithDefault("CACHE_EXPIRATION", "3600"), 10, 64)
	if err != nil {
		return &Config{}, err
	}

	ginMode := GetEnvWithDefault("GIN_MODE", "debug")
	appName := GetEnvWithDefault("APP_NAME", "dcard_homework")
	logger := newLogger(appName)

	return &Config{
		Port:    GetEnvWithDefault("PORT", "80"),
		GinMode: ginMode,
		Logger:  logger,
		RedisConfig: &RedisConfig{
			Addr:            os.Getenv("REDIS_ADDR"),
			Password:        GetEnvWithDefault("REDIS_PASSWD", ""),
			DB:              redisDB,
			PoolSize:        redisPoolSize,
			MaxRetries:      redisMaxRetries,
			IdleTimeout:     time.Duration(redisIdleTimeout) * time.Second,
			CacheExpiration: time.Duration(cacheExpiration) * time.Second,
		},
	}, nil
}
