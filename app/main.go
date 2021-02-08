package main

import log "github.com/sirupsen/logrus"

func main() {
	var err error
	var rateLimiterMiddleware *RateLimiterMiddleware
	rateLimiterMiddleware, err = InitializeRateLimiterMiddleware()
	if err != nil {
		log.Fatal(err)
	}
	ginMiddlewareCollection := NewGinMiddlewareCollection(
		rateLimiterMiddleware,
	)
	server, err := InitializeServer(ginMiddlewareCollection)
	if err != nil {
		log.Fatal(err)
	}
	server.Run()
}
