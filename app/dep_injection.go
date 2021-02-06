package main

import "go.uber.org/dig"

// BuildContainer is a factory for dependency injection (DI) container
func BuildContainer() *dig.Container {
	container := dig.New()

	container.Provide(NewConfig)
	container.Provide(NewRedisLimiterRepository)
	container.Provide(NewEngine)
	container.Provide(NewServer)

	return container
}
