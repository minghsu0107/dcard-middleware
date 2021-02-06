package main

import (
	"io"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Server is the http wrapper
type Server struct {
	Config *Config
	Engine *gin.Engine
}

// NewEngine is a factory for gin engine instance
// Global Middlewares and api log configurations are registered here
func NewEngine(config *Config, rateLimiterMiddleware *RateLimiterMiddleware) *gin.Engine {
	gin.SetMode(config.GinMode)
	if config.GinMode == "release" {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.DebugLevel)
	}
	gin.DefaultWriter = io.Writer(config.Logger.Writer)
	log.SetOutput(gin.DefaultWriter)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(rateLimiterMiddleware.RateLimiter())
	return engine
}

// NewServer is the factory for server instance
func NewServer(config *Config, engine *gin.Engine) *Server {
	return &Server{
		Config: config,
		Engine: engine,
	}
}

// RegisterRoutes method register all endpoints
func (s *Server) RegisterRoutes() {
	s.Engine.GET("/", helloworld)
}

// Run is a method for starting server
func (s *Server) Run() {
	s.RegisterRoutes()

	Addr := ":" + s.Config.Port
	s.Engine.Run(Addr)
}
