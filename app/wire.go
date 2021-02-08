//+build wireinject

package main

import "github.com/google/wire"

func InitializeRateLimiterMiddleware() (*RateLimiterMiddleware, error) {
	wire.Build(
		NewConfig,
		NewRedisLimiterRepository,
		NewRateLimiterMiddleware,
	)
	return &RateLimiterMiddleware{}, nil
}

func InitializeServer(ginMiddlewareCollection *GinMiddlewareCollection) (*Server, error) {
	wire.Build(
		NewConfig,
		NewEngine,
		NewServer,
	)
	return &Server{}, nil
}
