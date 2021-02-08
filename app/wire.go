//+build wireinject

package main

import "github.com/google/wire"

func InitializeServer() (*Server, error) {
	wire.Build(
		NewConfig,
		NewEngine,
		NewRedisLimiterRepository,
		NewRateLimiterMiddleware,
		NewServer,
	)
	return &Server{}, nil
}
